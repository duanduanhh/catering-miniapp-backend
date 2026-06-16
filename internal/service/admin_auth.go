package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

// ErrAdminLogin 表示账号或密码错误，由 handler 捕获后映射为 v1.ErrAdminLoginFailed
var ErrAdminLogin = errors.New("invalid admin username or password")

type AdminAuthService interface {
	Login(ctx context.Context, username, password string) (token string, expiresAt int64, err error)
}

func NewAdminAuthService(conf *viper.Viper) AdminAuthService {
	expireHours := conf.GetInt("admin.jwt_expire_hours")
	if expireHours <= 0 {
		expireHours = 24
	}
	return &adminAuthService{
		username:    conf.GetString("admin.username"),
		password:    conf.GetString("admin.password"),
		jwtSecret:   []byte(conf.GetString("admin.jwt_secret")),
		expireHours: expireHours,
	}
}

type adminAuthService struct {
	username    string
	password    string
	jwtSecret   []byte
	expireHours int
}

func (s *adminAuthService) Login(_ context.Context, username, password string) (string, int64, error) {
	// 配置缺失防呆
	if s.username == "" || s.password == "" || len(s.jwtSecret) == 0 {
		return "", 0, errors.New("admin credentials not configured")
	}
	if username != s.username || password != s.password {
		return "", 0, ErrAdminLogin
	}
	expiresAt := time.Now().Add(time.Duration(s.expireHours) * time.Hour)
	claims := middleware.AdminClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Subject:   username,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}
	return tokenString, expiresAt.Unix(), nil
}
