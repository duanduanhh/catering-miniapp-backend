package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/spf13/viper"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/pkg/jwt"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type WechatService interface {
	Register(ctx context.Context, code, loginCode string, inviterID int64) (string, *model.User, error)
	Login(ctx context.Context, code string) (token string, user *model.User, isOldUser bool, err error)
}

func NewWechatService(
	logger *log.Logger,
	config *viper.Viper,
	jwtClient *jwt.JWT,
	userRepo repository.UserRepository,
	contactVoucherHistoryRepo repository.ContactVoucherHistoryRepository,
) WechatService {
	return &wechatService{
		logger:                    logger,
		config:                    config,
		jwt:                       jwtClient,
		userRepo:                  userRepo,
		contactVoucherHistoryRepo: contactVoucherHistoryRepo,
	}
}

type wechatService struct {
	logger                    *log.Logger
	config                    *viper.Viper
	jwt                       *jwt.JWT
	userRepo                  repository.UserRepository
	contactVoucherHistoryRepo repository.ContactVoucherHistoryRepository
}

type wechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

type wechatAccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type wechatPhoneResponse struct {
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
	PhoneInfo struct {
		PurePhoneNumber string `json:"purePhoneNumber"`
	} `json:"phone_info"`
}

func (s *wechatService) Register(ctx context.Context, code, loginCode string, inviterID int64) (string, *model.User, error) {
	if code == "" || loginCode == "" {
		return "", nil, errors.New("code or loginCode is empty")
	}
	session, err := s.code2session(ctx, loginCode)
	if err != nil {
		return "", nil, err
	}
	phone, err := s.getPhone(ctx, code)
	if err != nil {
		return "", nil, err
	}
	user, err := s.userRepo.GetByOpenID(ctx, session.OpenID)
	if err != nil {
		return "", nil, err
	}
	now := time.Now()
	if user != nil {
		return "", nil, ErrUserExists
	}
	const newUserVoucherNum = 5
	var userCode string
	for i := 0; i < 5; i++ {
		candidate := randAlphanumeric(8)
		exists, err := s.userRepo.ExistsByUserCode(ctx, candidate)
		if err != nil {
			return "", nil, err
		}
		if !exists {
			userCode = candidate
			break
		}
	}
	if userCode == "" {
		return "", nil, errors.New("failed to generate unique user code")
	}
	user = &model.User{
		WechatOpenID:      session.OpenID,
		Phone:             phone,
		UserCode:          userCode,
		Name:              "餐饮人" + randAlphanumeric(6),
		InviterID:         inviterID,
		ContactVoucherNum: newUserVoucherNum,
		CreateAt:          now,
		UpdateAt:          now,
	}
	if err := s.userRepo.Create(ctx, user); err != nil {
		return "", nil, err
	}
	token, err := s.jwt.GenTokenWithOpenid(strconv.FormatInt(user.ID, 10), user.WechatOpenID, time.Now().Add(s.config.GetDuration("security.jwt.expire")))
	if err != nil {
		return "", nil, err
	}
	user.Token = token
	user.UpdateAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", nil, err
	}
	history := &model.ContactVoucherHistory{
		UserID:    user.ID,
		BizType:   model.ContactVoucherHistoryBuy,
		ChangeNum: newUserVoucherNum,
		LastNum:   0,
		NextNum:   newUserVoucherNum,
		Remark:    "新用户登陆",
		CreateAt:  time.Now(),
	}
	if err := s.contactVoucherHistoryRepo.Create(ctx, history); err != nil {
		return "", nil, err
	}
	if inviterID > 0 {
		// 累加邀请人的 invite_num
		inviter, err := s.userRepo.GetByID(ctx, inviterID)
		if err == nil {
			inviter.InviteNum++
			inviter.UpdateAt = time.Now()
			_ = s.userRepo.Update(ctx, inviter)
		}
		// 奖励邀请人 2 张联系券
		_ = rewardInviter(ctx, user.ID, 2, "邀请好友注册奖励", s.userRepo, s.contactVoucherHistoryRepo)
	}
	return token, user, nil
}

func (s *wechatService) Login(ctx context.Context, code string) (string, *model.User, bool, error) {
	if code == "" {
		return "", nil, false, errors.New("code is empty")
	}
	session, err := s.code2session(ctx, code)
	if err != nil {
		return "", nil, false, err
	}
	user, err := s.userRepo.GetByOpenID(ctx, session.OpenID)
	if err != nil {
		return "", nil, false, err
	}
	if user == nil {
		return "", nil, false, ErrUserNotFound
	}
	isOldUser := user.OldUserStatus == 1
	if isOldUser {
		user.OldUserStatus = 2
	}
	token, err := s.jwt.GenTokenWithOpenid(strconv.FormatInt(user.ID, 10), user.WechatOpenID, time.Now().Add(s.config.GetDuration("security.jwt.expire")))
	if err != nil {
		return "", nil, false, err
	}
	user.Token = token
	user.UpdateAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return "", nil, false, err
	}
	return token, user, isOldUser, nil
}

func randAlphanumeric(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

func (s *wechatService) code2session(ctx context.Context, code string) (wechatSessionResponse, error) {
	endpoint := s.config.GetString("wechat.endpoint_code2session")
	appID := s.config.GetString("wechat.app_id")
	secret := s.config.GetString("wechat.secret")
	if endpoint == "" || appID == "" || secret == "" {
		return wechatSessionResponse{}, errors.New("wechat config missing")
	}
	params := fmt.Sprintf("appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, secret, code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params, nil)
	if err != nil {
		return wechatSessionResponse{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return wechatSessionResponse{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wechatSessionResponse{}, err
	}
	var result wechatSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return wechatSessionResponse{}, err
	}
	if result.ErrCode != 0 {
		return wechatSessionResponse{}, fmt.Errorf("code2session error: %d %s", result.ErrCode, result.ErrMsg)
	}
	if result.OpenID == "" {
		return wechatSessionResponse{}, errors.New("openid is empty")
	}
	return result, nil
}

func (s *wechatService) getAccessToken(ctx context.Context) (wechatAccessToken, error) {
	endpoint := s.config.GetString("wechat.endpoint_access_token")
	appID := s.config.GetString("wechat.app_id")
	secret := s.config.GetString("wechat.secret")
	if endpoint == "" || appID == "" || secret == "" {
		return wechatAccessToken{}, errors.New("wechat config missing")
	}
	params := fmt.Sprintf("grant_type=client_credential&appid=%s&secret=%s", appID, secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params, nil)
	if err != nil {
		return wechatAccessToken{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return wechatAccessToken{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return wechatAccessToken{}, err
	}
	var result wechatAccessToken
	if err := json.Unmarshal(body, &result); err != nil {
		return wechatAccessToken{}, err
	}
	if result.ErrCode != 0 {
		return wechatAccessToken{}, fmt.Errorf("access_token error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result, nil
}

func (s *wechatService) getPhone(ctx context.Context, code string) (string, error) {
	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return "", err
	}
	endpoint := s.config.GetString("wechat.endpoint_phone")
	if endpoint == "" {
		return "", errors.New("wechat config missing")
	}
	payload := map[string]string{"code": code}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	url := endpoint + "?access_token=" + accessToken.AccessToken
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result wechatPhoneResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("phone error: %d %s", result.ErrCode, result.ErrMsg)
	}
	return result.PhoneInfo.PurePhoneNumber, nil
}
