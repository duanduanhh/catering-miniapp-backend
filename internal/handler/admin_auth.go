package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type AdminAuthHandler struct {
	*Handler
	adminAuthService service.AdminAuthService
}

func NewAdminAuthHandler(
	handler *Handler,
	adminAuthService service.AdminAuthService,
) *AdminAuthHandler {
	return &AdminAuthHandler{
		Handler:          handler,
		adminAuthService: adminAuthService,
	}
}

// Login godoc
// @Summary  管理后台登录
// @Description 校验管理员账号密码，签发 24 小时有效期的 JWT。Token 通过 `token` header 携带访问其它 /admin/* 接口。
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminLoginRequest true "params"
// @Success  200 {object} v1.AdminLoginResponseData
// @Router   /admin/login [post]
func (h *AdminAuthHandler) Login(ctx *gin.Context) {
	var req v1.AdminLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	token, expiresAt, err := h.adminAuthService.Login(ctx, req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrAdminLogin) {
			v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrAdminLoginFailed, "invalid username or password")
			return
		}
		h.logger.WithContext(ctx).Error("adminAuthService.Login error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.AdminLoginResponseData{
		Token:     token,
		ExpiresAt: expiresAt,
		Username:  req.Username,
	})
}
