package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/TangTangChu/AnzuImg/backend/internal/config"
	"github.com/TangTangChu/AnzuImg/backend/internal/logger"
)

var (
	ErrURLInvalid     = errors.New("invalid url")
	ErrURLBlocked     = errors.New("url blocked: target address not allowed")
	ErrURLTooLarge    = errors.New("url response exceeds size limit")
	ErrURLFetchFailed = errors.New("url fetch failed")
)

const maxFetchRedirects = 5

// URLFetcher 提供受控的 HTTP(S) 出站下载，专用于服务器端从用户提供的链接抓取图像。
// 默认拒绝任何指向私网/回环/链路本地/多播/未指定/Metadata 的目标 IP，
// 通过 net.Dialer.Control 在 TCP connect 前再次校验已解析 IP, 阻断 DNS rebinding。
type URLFetcher struct {
	cfg *config.Config
	log *logger.Logger
}

func NewURLFetcher(cfg *config.Config) *URLFetcher {
	return &URLFetcher{
		cfg: cfg,
		log: logger.Register("url-fetcher"),
	}
}

type FetchResult struct {
	Body     []byte
	MimeHint string
	FinalURL string
	Filename string
}

func (f *URLFetcher) Fetch(ctx context.Context, rawURL string) (*FetchResult, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLInvalid, err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("%w: unsupported scheme", ErrURLInvalid)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("%w: missing host", ErrURLInvalid)
	}

	eff := f.cfg.Effective()
	timeout := time.Duration(eff.URLFetchTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	maxBytes := eff.URLFetchMaxBytes
	if maxBytes <= 0 {
		maxBytes = 60 * 1024 * 1024
	}
	allowPrivate := eff.URLFetchAllowPrivate

	client := f.buildClient(timeout, allowPrivate)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLFetchFailed, err)
	}
	req.Header.Set("User-Agent", "AnzuImg-URLFetcher/1.0")
	req.Header.Set("Accept", "*/*")

	resp, err := client.Do(req)
	if err != nil {
		if isBlockedDialErr(err) {
			return nil, ErrURLBlocked
		}
		return nil, fmt.Errorf("%w: %v", ErrURLFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: status %d", ErrURLFetchFailed, resp.StatusCode)
	}
	if resp.ContentLength > 0 && resp.ContentLength > maxBytes {
		return nil, ErrURLTooLarge
	}

	limited := io.LimitReader(resp.Body, maxBytes+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrURLFetchFailed, err)
	}
	if int64(len(buf)) > maxBytes {
		return nil, ErrURLTooLarge
	}

	ctype := resp.Header.Get("Content-Type")
	if idx := strings.Index(ctype, ";"); idx >= 0 {
		ctype = strings.TrimSpace(ctype[:idx])
	}
	ctype = strings.ToLower(strings.TrimSpace(ctype))

	finalURL := parsed.String()
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}

	filename := derivePathBasename(resp.Request)
	if filename == "" {
		filename = derivePathBasename(req)
	}

	return &FetchResult{
		Body:     buf,
		MimeHint: ctype,
		FinalURL: finalURL,
		Filename: filename,
	}, nil
}

func (f *URLFetcher) buildClient(timeout time.Duration, allowPrivate bool) *http.Client {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
		Control: func(network, address string, _ syscall.RawConn) error {
			return verifyDialAddress(network, address, allowPrivate)
		},
	}
	transport := &http.Transport{
		DialContext:            dialer.DialContext,
		TLSHandshakeTimeout:    5 * time.Second,
		ResponseHeaderTimeout:  10 * time.Second,
		ExpectContinueTimeout:  1 * time.Second,
		DisableKeepAlives:      true,
		MaxIdleConns:           0,
		MaxResponseHeaderBytes: 1 << 16,
	}
	return &http.Client{
		Timeout:       timeout,
		Transport:     transport,
		CheckRedirect: checkRedirect,
	}
}

func checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxFetchRedirects {
		return fmt.Errorf("%w: too many redirects", ErrURLFetchFailed)
	}
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return fmt.Errorf("%w: redirect to unsupported scheme %s", ErrURLInvalid, req.URL.Scheme)
	}
	return nil
}

func verifyDialAddress(_ string, address string, allowPrivate bool) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return ErrURLBlocked
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return ErrURLBlocked
	}
	if allowPrivate {
		return nil
	}
	if !isPublicIP(ip) {
		return ErrURLBlocked
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsUnspecified() {
		return false
	}
	if ip.IsLoopback() {
		return false
	}
	if ip.IsPrivate() {
		return false
	}
	if ip.IsLinkLocalUnicast() {
		return false
	}
	if ip.IsLinkLocalMulticast() {
		return false
	}
	if ip.IsMulticast() {
		return false
	}
	if ip.IsInterfaceLocalMulticast() {
		return false
	}
	if v4 := ip.To4(); v4 != nil {
		switch v4[0] {
		case 0:
			return false
		case 100:
			if v4[1] >= 64 && v4[1] <= 127 {
				return false
			}
		case 198:
			if v4[1] == 18 || v4[1] == 19 {
				return false
			}
		}
	}
	return true
}

func isBlockedDialErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrURLBlocked) {
		return true
	}
	return strings.Contains(err.Error(), ErrURLBlocked.Error())
}

func derivePathBasename(req *http.Request) string {
	if req == nil || req.URL == nil {
		return ""
	}
	p := req.URL.Path
	if p == "" {
		return ""
	}
	base := path.Base(p)
	if base == "." || base == "/" || base == "" {
		return ""
	}
	if decoded, err := url.QueryUnescape(base); err == nil {
		return decoded
	}
	return base
}
