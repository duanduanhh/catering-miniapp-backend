package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"github.com/go-nunu/nunu-layout-advanced/pkg/jwt"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
)

type Handler struct {
	logger *log.Logger
}

func isPaymentPackageOrderError(err error) bool {
	return errors.Is(err, service.ErrPaymentPackageNotFound) ||
		errors.Is(err, service.ErrPaymentPackageInvalid) ||
		errors.Is(err, service.ErrPaymentPackageUnavailable) ||
		errors.Is(err, service.ErrPaymentPackageLimitReached)
}

func NewHandler(
	logger *log.Logger,
) *Handler {
	return &Handler{
		logger: logger,
	}
}

func GetUserIdFromCtx(ctx *gin.Context) int64 {
	v, exists := ctx.Get("claims")
	if !exists {
		return 0
	}
	userID := v.(*jwt.MyCustomClaims).UserId
	parsed, err := strconv.ParseInt(userID, 10, 64)
	if err != nil || parsed <= 0 {
		return 0
	}
	return parsed
}

func GetOpenidFromCtx(ctx *gin.Context) string {
	v, exists := ctx.Get("claims")
	if !exists {
		return ""
	}
	return v.(*jwt.MyCustomClaims).Openid
}
