package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
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
	// 微信图片审核和审核前压缩暂时关闭，保留相关实现以便后续恢复。
	// check_mode 当前不生效，原始图片直接上传 OSS。
	ext := strings.ToLower(filepath.Ext(originalFilename))
	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	url, err := s.upload(ctx, file, filename)
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
