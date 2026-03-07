package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cshum/vipsgen/vips"
)

// ImageInfo 图片信息结构体
type ImageInfo struct {
	MimeType string // MIME类型
	Width    int    // 图片宽度
	Height   int    // 图片高度
	Size     int64  // 文件大小（字节）
}

var formatToMime = map[vips.ImageType]string{
	vips.ImageTypeJpeg:    "image/jpeg",
	vips.ImageTypePng:     "image/png",
	vips.ImageTypeWebp:    "image/webp",
	vips.ImageTypeGif:     "image/gif",
	vips.ImageTypeSvg:     "image/svg+xml",
	vips.ImageTypeHeif:    "image/heif",
	vips.ImageTypeAvif:    "image/avif",
	vips.ImageTypeTiff:    "image/tiff",
	vips.ImageTypeBmp:     "image/bmp",
	vips.ImageTypeJxl:     "image/jxl",
	vips.ImageTypeUnknown: "application/octet-stream",
}

func detectAVIFBrand(reader io.Reader) bool {
	buf := make([]byte, 12)
	n, err := io.ReadFull(reader, buf)
	if err != nil || n < 12 {
		return false
	}
	if string(buf[4:8]) != "ftyp" {
		return false
	}
	majorBrand := string(buf[8:12])
	return majorBrand == "avif" || majorBrand == "avis" // avis 是 animated AVIF
}

// InspectImage 获取图片的宽高和 MIME 类型
func InspectImage(reader io.Reader) (mime string, width, height int, err error) {
	tee := io.TeeReader(reader, &bytes.Buffer{})

	source := vips.NewSource(io.NopCloser(tee))
	img, err := vips.NewImageFromSource(source, nil)
	if err != nil {
		return "", 0, 0, fmt.Errorf("failed to load image header: %v", err)
	}
	defer img.Close()

	width = img.Width()
	height = img.Height()

	format := img.Format()
	mime, ok := formatToMime[format]
	if !ok {
		mime = "application/octet-stream"
	}

	// 如果是 HEIF，进一步检查 brand
	if format == vips.ImageTypeHeif {
		if seeker, ok := reader.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		if detectAVIFBrand(reader) {
			mime = "image/avif"
		} else {
			mime = "image/heif"
		}
	}

	return mime, width, height, nil
}

// InspectImageWithInfo 获取完整的图片信息
func InspectImageWithInfo(reader io.Reader, dataSize int64) (*ImageInfo, error) {
	mime, width, height, err := InspectImage(reader)
	if err != nil {
		return nil, err
	}

	return &ImageInfo{
		MimeType: mime,
		Width:    width,
		Height:   height,
		Size:     dataSize,
	}, nil
}

// DetectMIMEType 检测字节数据的MIME类型
func DetectMIMEType(data []byte) string {
	mime, _, _, err := InspectImage(bytes.NewReader(data))
	if err != nil {
		return "application/octet-stream"
	}
	return mime
}

// IsImageFile 检查MIME类型是否为图片
func IsImageFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// IsVideoFile 检查MIME类型是否为视频
func IsVideoFile(mimeType string) bool {
	return strings.HasPrefix(mimeType, "video/")
}

// DetectImageDimensions 检测图片尺寸
func DetectImageDimensions(data []byte) (width, height int, err error) {
	_, width, height, err = InspectImage(bytes.NewReader(data))
	return width, height, err
}

// GenerateThumbnail 生成 WebP 缩略图
func GenerateThumbnail(reader io.Reader, width, height int) ([]byte, error) {
	source := vips.NewSource(io.NopCloser(reader))
	img, err := vips.NewThumbnailSource(source, width, &vips.ThumbnailSourceOptions{
		Height: height,
		Size:   vips.SizeDown,
	})
	if err != nil {
		return nil, fmt.Errorf("thumbnail failed: %v", err)
	}
	defer img.Close()
	buf, err := img.WebpsaveBuffer(&vips.WebpsaveBufferOptions{
		Q:      75,
		Effort: 4,
	})
	if err != nil {
		return nil, fmt.Errorf("webp encode failed: %v", err)
	}

	return buf, nil
}

// ConvertImage 将图片转换为指定格式
func ConvertImage(data []byte, sourceMimeType string, targetFormat string, quality int, effort int) ([]byte, string, error) {
	// 验证目标格式
	targetFormat = strings.ToLower(targetFormat)
	if targetFormat != "webp" && targetFormat != "avif" {
		return nil, "", fmt.Errorf("unsupported target format: %s", targetFormat)
	}

	// 设置默认值
	if quality <= 0 {
		if targetFormat == "webp" {
			quality = 80
		} else {
			quality = 50
		}
	}
	if effort <= 0 {
		effort = 4
	}

	img, err := loadForConversion(data, sourceMimeType)
	if err != nil {
		return nil, "", err
	}
	defer img.Close()

	pageHeight := 0
	isAnimated := false
	if img.Pages() > 1 {
		isAnimated = true
		pageHeight = img.PageHeight()
		if pageHeight <= 0 {
			pageHeight = img.Height()
		}
	}

	var buf []byte
	var mimeType string

	switch targetFormat {
	case "webp":
		buf, err = img.WebpsaveBuffer(&vips.WebpsaveBufferOptions{
			Q:          quality,
			Effort:     effort,
			PageHeight: pageHeight,
		})
		mimeType = "image/webp"
	case "avif":
		if isAnimated {
			if avifBuf, convErr := ConvertAnimatedToAvif(data, sourceMimeType, quality, effort); convErr == nil {
				return avifBuf, "image/avif", nil
			}
		}
		buf, err = img.HeifsaveBuffer(&vips.HeifsaveBufferOptions{
			Q:           quality,
			Effort:      effort,
			Compression: vips.HeifCompressionAv1,
			PageHeight:  pageHeight,
		})
		mimeType = "image/avif"
	}

	if err != nil {
		return nil, "", fmt.Errorf("convert failed: %v", err)
	}

	return buf, mimeType, nil
}

func loadForConversion(data []byte, sourceMimeType string) (*vips.Image, error) {
	switch {
	case sourceMimeType == "image/gif":
		img, err := vips.NewGifloadBuffer(data, &vips.GifloadBufferOptions{N: -1})
		if err != nil {
			return nil, fmt.Errorf("failed to load gif: %v", err)
		}
		return img, nil
	case sourceMimeType == "image/webp":
		img, err := vips.NewWebploadBuffer(data, &vips.WebploadBufferOptions{N: -1, Scale: 1})
		if err != nil {
			return nil, fmt.Errorf("failed to load webp: %v", err)
		}
		return img, nil
	case sourceMimeType == "image/heif" || sourceMimeType == "image/heic" || sourceMimeType == "image/avif":
		img, err := vips.NewHeifloadBuffer(data, &vips.HeifloadBufferOptions{N: -1})
		if err != nil {
			return nil, fmt.Errorf("failed to load heif/avif: %v", err)
		}
		return img, nil
	default:
		source := vips.NewSource(io.NopCloser(bytes.NewReader(data)))
		img, err := vips.NewImageFromSource(source, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to load image: %v", err)
		}
		return img, nil
	}
}

func ConvertAnimatedToAvif(data []byte, sourceMimeType string, quality, effort int) ([]byte, error) {
	inputExt := mimeTypeToExt(sourceMimeType)
	if inputExt == "" {
		inputExt = ".img"
	}

	inFile, err := os.CreateTemp("", "anzuimg-anim-*"+inputExt)
	if err != nil {
		return nil, fmt.Errorf("create temp input failed: %w", err)
	}
	inPath := inFile.Name()
	if _, err := inFile.Write(data); err != nil {
		_ = inFile.Close()
		_ = os.Remove(inPath)
		return nil, fmt.Errorf("write temp input failed: %w", err)
	}
	_ = inFile.Close()
	defer func() { _ = os.Remove(inPath) }()

	outFile, err := os.CreateTemp("", "anzuimg-anim-*.avif")
	if err != nil {
		return nil, fmt.Errorf("create temp output failed: %w", err)
	}
	outPath := outFile.Name()
	_ = outFile.Close()
	defer func() { _ = os.Remove(outPath) }()

	crf := qualityToAV1CRF(quality)
	cpuUsed := effortToAV1CPUUsed(effort)

	args := []string{
		"-y",
		"-i", inPath,
		"-c:v", "libaom-av1",
		"-crf", strconv.Itoa(crf),
		"-b:v", "0",
		"-cpu-used", strconv.Itoa(cpuUsed),
		"-still-picture", "0",
		"-row-mt", "1",
		"-pix_fmt", "yuv420p",
		"-f", "avif",
		outPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return nil, fmt.Errorf("ffmpeg not found")
		}
		return nil, fmt.Errorf("ffmpeg avif conversion failed: %w, output: %s", err, string(out))
	}

	buf, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read avif output failed: %w", err)
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty avif output")
	}

	return buf, nil
}

func qualityToAV1CRF(quality int) int {
	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}
	return 63 - int(float64(quality-1)*63.0/99.0)
}

func effortToAV1CPUUsed(effort int) int {
	if effort < 0 {
		effort = 0
	}
	if effort > 8 {
		effort = 8
	}
	return 8 - effort
}

func mimeTypeToExt(mimeType string) string {
	switch mimeType {
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/avif":
		return ".avif"
	case "image/heif", "image/heic":
		return ".heif"
	case "image/png":
		return ".png"
	case "image/jpeg", "image/jpg":
		return ".jpg"
	default:
		return ""
	}
}
