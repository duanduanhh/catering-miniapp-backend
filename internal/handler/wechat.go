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
	orderService service.OrderService
}

func NewWechatHandler(handler *Handler, orderService service.OrderService) *WechatHandler {
	return &WechatHandler{
		Handler:      handler,
		orderService: orderService,
	}
}

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
