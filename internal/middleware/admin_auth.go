package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

// AdminClaims 管理后台专用 JWT claims，与小程序用户 jwt 完全隔离
type AdminClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AdminAuth 校验 token header 是否为合法的 admin JWT
// 通过则把 *AdminClaims 存到 ctx 的 "admin_claims" 中。
func AdminAuth(conf *viper.Viper, logger *log.Logger) gin.HandlerFunc {
	secret := []byte(conf.GetString("admin.jwt_secret"))
	return func(ctx *gin.Context) {
		tokenString := strings.TrimSpace(ctx.Request.Header.Get("token"))
		if tokenString == "" {
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrAdminUnauthorized, "missing admin token")
			ctx.Abort()
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		claims := &AdminClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return secret, nil
		})
		if err != nil || token == nil || !token.Valid {
			logger.WithContext(ctx).Warn("admin token parse failed", zap.Error(err))
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrAdminUnauthorized, "invalid admin token")
			ctx.Abort()
			return
		}
		if strings.TrimSpace(claims.Username) == "" {
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrAdminUnauthorized, "empty admin username")
			ctx.Abort()
			return
		}

		ctx.Set("admin_claims", claims)
		ctx.Next()
	}
}

// GetAdminUsernameFromCtx 从 gin.Context 中取出当前管理员用户名，无则返回空串
func GetAdminUsernameFromCtx(ctx *gin.Context) string {
	v, exists := ctx.Get("admin_claims")
	if !exists {
		return ""
	}
	c, ok := v.(*AdminClaims)
	if !ok || c == nil {
		return ""
	}
	return c.Username
}
