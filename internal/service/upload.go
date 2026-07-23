package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

const (
	ImageCheckModeSync  = "sync"
	ImageCheckModeAsync = "async"

	ImageAuditStatusPassed  = "passed"
	ImageAuditStatusPending = "pending"

	maxWechatImageSecCheckSize   = 1024 * 1024
	maxWechatImageSecCheckWidth  = 750
	maxWechatImageSecCheckHeight = 1334
)

type UploadService interface {
	UploadImage(ctx context.Context, file io.Reader, filename string, userID int64, openid string, checkMode string) (*UploadImageResult, error)
}

type UploadImageResult struct {
	URL         string
	CheckMode   string
	AuditStatus string
	TraceID     string
}

type uploadService struct {
	config               *viper.Viper
	imageSecurityService ImageSecurityService
	client               *oss.Client
	bucket               *oss.Bucket
	bucketName           string
	urlPrefix            string
	mu                   sync.Mutex
}

func NewUploadService(config *viper.Viper, imageSecurityService ImageSecurityService) UploadService {
	return &uploadService{
		config:               config,
		imageSecurityService: imageSecurityService,
	}
}

func (s *uploadService) UploadImage(ctx context.Context, file io.Reader, originalFilename string, userID int64, openid string, checkMode string) (*UploadImageResult, error) {
	imageData, ext, err := compressForSecurityCheck(file)
	if err != nil {
		return nil, err
	}

	// 微信图片审核暂时关闭，保留同步/异步审核实现以便后续恢复。
	// check_mode 当前不生效，图片压缩后直接上传 OSS。
	url, err := s.uploadCompressedImage(ctx, imageData, ext, userID)
	if err != nil {
		return nil, err
	}
	return &UploadImageResult{URL: url}, nil
}

func (s *uploadService) uploadImageWithSyncCheck(ctx context.Context, imageData []byte, ext, originalFilename string, userID int64) (*UploadImageResult, error) {
	if err := s.imageSecurityService.CheckImage(ctx, originalFilename, imageData); err != nil {
		return nil, err
	}
	url, err := s.uploadCompressedImage(ctx, imageData, ext, userID)
	if err != nil {
		return nil, err
	}
	return &UploadImageResult{URL: url, CheckMode: ImageCheckModeSync, AuditStatus: ImageAuditStatusPassed}, nil
}

func (s *uploadService) uploadImageWithAsyncCheck(ctx context.Context, imageData []byte, ext string, userID int64, openid string) (*UploadImageResult, error) {
	url, err := s.uploadCompressedImage(ctx, imageData, ext, userID)
	if err != nil {
		return nil, err
	}
	traceID, err := s.imageSecurityService.CheckImageAsync(ctx, openid, url)
	if err != nil {
		return nil, err
	}
	return &UploadImageResult{URL: url, CheckMode: ImageCheckModeAsync, AuditStatus: ImageAuditStatusPending, TraceID: traceID}, nil
}

func (s *uploadService) uploadCompressedImage(ctx context.Context, imageData []byte, ext string, userID int64) (string, error) {
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	return s.upload(ctx, bytes.NewReader(imageData), filename)
}

func normalizeImageCheckMode(checkMode string) string {
	checkMode = strings.ToLower(strings.TrimSpace(checkMode))
	if checkMode == "" {
		return ImageCheckModeSync
	}
	return checkMode
}

func (s *uploadService) upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	bucket, err := s.ensureClient()
	if err != nil {
		return "", err
	}
	month := time.Now().Format("2006-01")
	objectKey := fmt.Sprintf("img/%s/%s", month, filename)
	if err := bucket.PutObject(objectKey, file); err != nil {
		return "", err
	}
	urlPrefix := s.urlPrefix
	if urlPrefix == "" {
		return "", errors.New("oss.url_prefix is empty")
	}
	if !strings.HasSuffix(urlPrefix, "/") {
		urlPrefix += "/"
	}
	return urlPrefix + month + "/" + filename, nil
}

func compressForSecurityCheck(file io.Reader) ([]byte, string, error) {
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, "", err
	}
	switch format {
	case "jpeg", "png", "gif":
	default:
		return nil, "", errors.New("unsupported image format")
	}
	return encodeJPEGWithinLimit(resizeWithinWechatLimit(img))
}

func resizeWithinWechatLimit(img image.Image) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxWechatImageSecCheckWidth && height <= maxWechatImageSecCheckHeight {
		return img
	}
	widthRatio := float64(maxWechatImageSecCheckWidth) / float64(width)
	heightRatio := float64(maxWechatImageSecCheckHeight) / float64(height)
	ratio := widthRatio
	if heightRatio < ratio {
		ratio = heightRatio
	}
	newWidth := int(float64(width) * ratio)
	newHeight := int(float64(height) * ratio)
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}
	return resizeNearest(img, newWidth, newHeight)
}

func encodeJPEGWithinLimit(img image.Image) ([]byte, string, error) {
	quality := 85
	for {
		data, err := encodeJPEG(img, quality)
		if err != nil {
			return nil, "", err
		}
		if len(data) <= maxWechatImageSecCheckSize {
			return data, ".jpg", nil
		}
		if quality > 45 {
			quality -= 10
			continue
		}
		bounds := img.Bounds()
		width := bounds.Dx()
		height := bounds.Dy()
		if width <= 1 || height <= 1 {
			return nil, "", errors.New("image cannot be compressed under 1MB")
		}
		newWidth := width * 85 / 100
		newHeight := height * 85 / 100
		if newWidth < 1 {
			newWidth = 1
		}
		if newHeight < 1 {
			newHeight = 1
		}
		img = resizeNearest(img, newWidth, newHeight)
		quality = 85
	}
}

func encodeJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, flattenImage(img), &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func flattenImage(src image.Image) image.Image {
	bounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(dst, dst.Bounds(), image.White, image.Point{}, draw.Src)
	draw.Draw(dst, dst.Bounds(), src, bounds.Min, draw.Over)
	return dst
}

func resizeNearest(src image.Image, width, height int) image.Image {
	srcBounds := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		sy := srcBounds.Min.Y + y*srcBounds.Dy()/height
		for x := 0; x < width; x++ {
			sx := srcBounds.Min.X + x*srcBounds.Dx()/width
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func (s *uploadService) ensureClient() (*oss.Bucket, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bucket != nil {
		return s.bucket, nil
	}
	accessKey := s.config.GetString("oss.access_key_id")
	secret := s.config.GetString("oss.access_key_secret")
	endpoint := s.config.GetString("oss.endpoint")
	bucketName := s.config.GetString("oss.bucket")
	urlPrefix := s.config.GetString("oss.url_prefix")
	if accessKey == "" || secret == "" || endpoint == "" || bucketName == "" {
		return nil, errors.New("oss config is incomplete")
	}
	client, err := oss.New(endpoint, accessKey, secret)
	if err != nil {
		return nil, err
	}
	ossBucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	s.client = client
	s.bucket = ossBucket
	s.bucketName = bucketName
	s.urlPrefix = urlPrefix
	return s.bucket, nil
}

func init() {
	image.RegisterFormat("jpeg", "\xff\xd8", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "\x89PNG\r\n\x1a\n", png.Decode, png.DecodeConfig)
	image.RegisterFormat("gif", "GIF8?a", gif.Decode, gif.DecodeConfig)
}
