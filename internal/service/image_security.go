package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

const wechatRiskyContentCode = 87014

type ImageSecurityService interface {
	CheckImage(ctx context.Context, filename string, data []byte) error
	CheckImageAsync(ctx context.Context, openid, mediaURL string) (string, error)
}

type imageSecurityService struct {
	config         *viper.Viper
	logger         *log.Logger
	mu             sync.Mutex
	accessToken    string
	tokenExpiresAt time.Time
}

type wechatImageSecCheckResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

type wechatMediaCheckAsyncRequest struct {
	OpenID    string `json:"openid"`
	Scene     int    `json:"scene"`
	Version   int    `json:"version"`
	MediaURL  string `json:"media_url"`
	MediaType int    `json:"media_type"`
}

type wechatMediaCheckAsyncResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	TraceID string `json:"trace_id"`
}

func NewImageSecurityService(config *viper.Viper, logger *log.Logger) ImageSecurityService {
	return &imageSecurityService{
		config: config,
		logger: logger,
	}
}

func (s *imageSecurityService) CheckImage(ctx context.Context, filename string, data []byte) error {
	if len(data) == 0 {
		return errors.New("image data is empty")
	}
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return err
	}
	endpoint := s.config.GetString("wechat.endpoint_img_sec_check")
	if endpoint == "" {
		endpoint = "https://api.weixin.qq.com/wxa/img_sec_check"
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return err
	}
	if _, err := part.Write(data); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?access_token="+accessToken, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result wechatImageSecCheckResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return err
	}
	s.logger.WithContext(ctx).Info("wechat img_sec_check response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("filename", filename),
		zap.Int("image_size", len(data)),
		zap.Int("errcode", result.ErrCode),
		zap.String("errmsg", result.ErrMsg),
		zap.ByteString("response_body", respBody),
	)
	if result.ErrCode == 0 {
		return nil
	}
	if result.ErrCode == wechatRiskyContentCode {
		return ErrImageRiskyContent
	}
	s.logger.WithContext(ctx).Error("wechat img_sec_check error", zap.Int("errcode", result.ErrCode), zap.String("errmsg", result.ErrMsg))
	return fmt.Errorf("img_sec_check error: %d %s", result.ErrCode, result.ErrMsg)
}

func (s *imageSecurityService) CheckImageAsync(ctx context.Context, openid, mediaURL string) (string, error) {
	if openid == "" {
		return "", errors.New("openid is empty")
	}
	if mediaURL == "" {
		return "", errors.New("media_url is empty")
	}
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return "", err
	}
	endpoint := s.config.GetString("wechat.endpoint_media_check_async")
	if endpoint == "" {
		endpoint = "https://api.weixin.qq.com/wxa/media_check_async"
	}
	payload := wechatMediaCheckAsyncRequest{
		OpenID:    openid,
		Scene:     1,
		Version:   2,
		MediaURL:  mediaURL,
		MediaType: 2,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?access_token="+accessToken, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result wechatMediaCheckAsyncResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	s.logger.WithContext(ctx).Info("wechat media_check_async response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("media_url", mediaURL),
		zap.Int("errcode", result.ErrCode),
		zap.String("errmsg", result.ErrMsg),
		zap.String("trace_id", result.TraceID),
		zap.ByteString("response_body", respBody),
	)
	if result.ErrCode != 0 {
		return "", fmt.Errorf("media_check_async error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result.TraceID, nil
}

func (s *imageSecurityService) getAccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	if s.accessToken != "" && time.Now().Before(s.tokenExpiresAt) {
		accessToken := s.accessToken
		s.mu.Unlock()
		return accessToken, nil
	}
	s.mu.Unlock()

	token, expiresIn, err := s.fetchAccessToken(ctx)
	if err != nil {
		return "", err
	}
	if expiresIn <= 0 {
		expiresIn = 7200
	}

	s.mu.Lock()
	s.accessToken = token
	s.tokenExpiresAt = time.Now().Add(time.Duration(expiresIn-300) * time.Second)
	s.mu.Unlock()
	return token, nil
}

func (s *imageSecurityService) fetchAccessToken(ctx context.Context) (string, int64, error) {
	endpoint := s.config.GetString("wechat.endpoint_access_token")
	appID := s.config.GetString("wechat.app_id")
	secret := s.config.GetString("wechat.secret")
	if endpoint == "" || appID == "" || secret == "" {
		return "", 0, errors.New("wechat config missing")
	}
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", endpoint, appID, secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	var result wechatAccessToken
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, err
	}
	if result.ErrCode != 0 {
		return "", 0, fmt.Errorf("access_token error: %d %s", result.ErrCode, result.ErrMsg)
	}
	if result.AccessToken == "" {
		return "", 0, errors.New("access_token is empty")
	}
	return result.AccessToken, result.ExpiresIn, nil
}
