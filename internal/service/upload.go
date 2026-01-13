package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

type UploadService interface {
	Upload(ctx context.Context, file io.Reader, filename string) (string, error)
}

type uploadService struct {
	config     *viper.Viper
	client     *oss.Client
	bucket     *oss.Bucket
	bucketName string
	urlPrefix  string
	mu         sync.Mutex
}

func NewUploadService(config *viper.Viper) UploadService {
	return &uploadService{
		config: config,
	}
}

func (s *uploadService) Upload(ctx context.Context, file io.Reader, filename string) (string, error) {
	bucket, err := s.ensureClient()
	if err != nil {
		return "", err
	}
	objectKey := fmt.Sprintf("img/%s", filename)
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
	return urlPrefix + filename, nil
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
