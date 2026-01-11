package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/pkg/jwt"
	"github.com/go-nunu/nunu-layout-advanced/pkg/log"
	"strconv"
)

type Handler struct {
	logger *log.Logger
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
		return getUserIdFromHeader(ctx)
	}
	userID := v.(*jwt.MyCustomClaims).UserId
	parsed, err := strconv.ParseInt(userID, 10, 64)
	if err != nil || parsed <= 0 {
		return getUserIdFromHeader(ctx)
	}
	return parsed
}

func getUserIdFromHeader(ctx *gin.Context) int64 {
	userID := ctx.GetHeader("user_id")
	if userID == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
