package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
)

type WechatHandler struct {
	*Handler
	orderService  service.OrderService
	wechatService service.WechatService
}

func NewWechatHandler(handler *Handler, orderService service.OrderService, wechatService service.WechatService) *WechatHandler {
	return &WechatHandler{
		Handler:       handler,
		orderService:  orderService,
		wechatService: wechatService,
	}
}

// Register godoc
// @Summary 微信注册
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body v1.WechatRegisterRequest true "params"
// @Success 200 {object} v1.WechatLoginResponseData
// @Router /wechat/user/register [post]
func (h *WechatHandler) Register(ctx *gin.Context) {
	var req v1.WechatRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	token, user, err := h.wechatService.Register(ctx, req.PhoneCode, req.LoginCode, req.InviterID)
	if err != nil {
		h.logger.WithContext(ctx).Error("wechatService.Register error", zap.Error(err))
		if err == service.ErrUserExists {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	ctx.Header("Authorization", "Bearer "+token)
	v1.HandleSuccess(ctx, v1.WechatLoginResponseData{
		UserInfo: v1.WechatLoginUserInfo{ID: user.ID},
	})
}

// Login godoc
// @Summary 微信登录
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body v1.WechatLoginRequest true "params"
// @Success 200 {object} v1.WechatLoginResponseData
// @Router /wechat/user/login [post]
func (h *WechatHandler) Login(ctx *gin.Context) {
	var req v1.WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	token, user, err := h.wechatService.Login(ctx, req.LoginCode)
	if err != nil {
		h.logger.WithContext(ctx).Error("wechatService.Login error", zap.Error(err))
		if err == service.ErrUserNotFound {
			v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	ctx.Header("Authorization", "Bearer "+token)
	v1.HandleSuccess(ctx, v1.WechatLoginResponseData{
		UserInfo: v1.WechatLoginUserInfo{ID: user.ID},
	})
}

// Pay godoc
// @Summary 微信支付
// @Tags 支付模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.WechatPayRequest true "params"
// @Success 200 {object} v1.Response
// @Router /wechat/pay [post]
func (h *WechatHandler) Pay(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.WechatPayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	_, err := h.orderService.PayOrder(ctx, userID, req.OrderID, req.OrderNo, req.Price, "wxpay", "wxpay_"+time.Now().Format("20060102150405"))
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.PayOrder error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		if err == service.ErrAmountMismatch {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrAmountMismatch, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}
