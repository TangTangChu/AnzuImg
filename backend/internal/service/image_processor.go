package service

import (
	"bytes"
	"fmt"
	"io"
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
func ConvertImage(reader io.Reader, targetFormat string, quality int, effort int) ([]byte, string, error) {
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

	source := vips.NewSource(io.NopCloser(reader))
	img, err := vips.NewImageFromSource(source, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load image: %v", err)
	}
	defer img.Close()

	var buf []byte
	var mimeType string

	switch targetFormat {
	case "webp":
		buf, err = img.WebpsaveBuffer(&vips.WebpsaveBufferOptions{
			Q:      quality,
			Effort: effort,
		})
		mimeType = "image/webp"
	case "avif":
		buf, err = img.HeifsaveBuffer(&vips.HeifsaveBufferOptions{
			Q:           quality,
			Effort:      effort,
			Compression: vips.HeifCompressionAv1,
		})
		mimeType = "image/avif"
	}

	if err != nil {
		return nil, "", fmt.Errorf("convert failed: %v", err)
	}

	return buf, mimeType, nil
}

func mapLoaderToMime(loader string) string {
	loader = strings.ToLower(loader)
	if strings.Contains(loader, "jpeg") {
		return "image/jpeg"
	}
	if strings.Contains(loader, "png") {
		return "image/png"
	}
	if strings.Contains(loader, "webp") {
		return "image/webp"
	}
	if strings.Contains(loader, "gif") {
		return "image/gif"
	}
	if strings.Contains(loader, "heif") {
		return "image/avif"
	}
	if strings.Contains(loader, "svg") {
		return "image/svg+xml"
	}
	if strings.Contains(loader, "jxl") {
		return "image/jxl"
	}
	return "application/octet-stream"
}
