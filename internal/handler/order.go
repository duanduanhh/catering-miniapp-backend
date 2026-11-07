package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type OrderHandler struct {
	*Handler
	orderService          service.OrderService
	virtualPaymentService service.VirtualPaymentService
	wechatService         service.WechatService
}

func NewOrderHandler(
	handler *Handler,
	orderService service.OrderService,
	virtualPaymentService service.VirtualPaymentService,
	wechatService service.WechatService,
) *OrderHandler {
	return &OrderHandler{
		Handler:               handler,
		orderService:          orderService,
		virtualPaymentService: virtualPaymentService,
		wechatService:         wechatService,
	}
}

// PrepareVirtualPayment godoc
// @Summary 准备微信虚拟支付
// @Description 为当前用户的待支付单生成 wx.requestVirtualPayment 所需的 signData、paySig、signature。login_code 必须在本次支付前由 wx.login() 新获取；后端校验其所属用户与订单归属。小程序必须原样传入 virtual_payment，不能修改或重新序列化 signData。
// @Tags 支付模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.VirtualPaymentPrepareRequest true "订单号和本次 wx.login code"
// @Success 200 {object} v1.Response{data=v1.VirtualPaymentPrepareResponseData} "签名生成成功"
// @Failure 400 {object} v1.Response "订单不可支付、SKU 未配置微信道具 ID 或虚拟支付未配置"
// @Failure 401 {object} v1.Response "未登录或 login_code 不属于当前用户"
// @Router /payment/virtual/prepare [post]
func (h *OrderHandler) PrepareVirtualPayment(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	openID := GetOpenidFromCtx(ctx)
	if userID == 0 || openID == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.VirtualPaymentPrepareRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	order, err := h.orderService.GetVirtualPaymentOrder(ctx, userID, req.OrderNo)
	if err != nil {
		h.handleVirtualPaymentError(ctx, err)
		return
	}
	sessionKey, err := h.wechatService.SessionKeyForPayment(ctx, req.LoginCode, openID)
	if err != nil {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, err.Error())
		return
	}
	params, err := h.virtualPaymentService.Prepare(ctx, order, sessionKey)
	if err != nil {
		h.handleVirtualPaymentError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, v1.VirtualPaymentPrepareResponseData{
		OrderNo:        order.OrderNo,
		AmountCents:    order.AmountCents,
		VirtualPayment: params,
	})
}

func (h *OrderHandler) handleVirtualPaymentError(ctx *gin.Context, err error) {
	h.logger.WithContext(ctx).Error("virtual payment prepare error", zap.Error(err))
	if errors.Is(err, service.ErrForbidden) {
		v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
		return
	}
	if errors.Is(err, service.ErrVirtualPaymentUnavailable) || errors.Is(err, service.ErrVirtualPaymentOrderInvalid) || errors.Is(err, service.ErrVirtualPaymentOrderClosed) || errors.Is(err, service.ErrPaymentPackageNotFound) {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
}

// QueryOrderStatus godoc
// @Summary 查询订单支付状态
// @Description 前端唤起微信支付后，通过此接口确认订单在后端的真实支付状态。status: 1=待支付 2=已支付 3=已取消 4=已退款。建议支付完成后轮询直到 status=2 再刷新业务页面。
// @Tags 订单
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.OrderStatusRequest true "params"
// @Success 200 {object} v1.OrderStatusResponseData
// @Router /order/status [post]
func (h *OrderHandler) QueryOrderStatus(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.OrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	order, err := h.orderService.GetByOrderNo(ctx, userID, req.OrderNo)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.GetByOrderNo error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.OrderStatusResponseData{
		OrderNo: order.OrderNo,
		Status:  int(order.Status),
	})
}

// ListOrders godoc
// @Summary 消费记录
// @Description 返回当前用户已支付的订单列表，按创建时间倒序。product_type: 1=岗位置顶 2=联系券 3=付费刷新 4=招租发布。amount 单位：元。
// @Tags 个人中心
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.UserOrderListRequest true "params"
// @Success 200 {object} v1.UserOrderListResponseData
// @Router /user/orders [post]
func (h *OrderHandler) ListOrders(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.UserOrderListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	list, total, err := h.orderService.ListByUser(ctx, userID, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.UserOrderListResponseData{
		List:  make([]v1.UserOrderListItem, 0, len(list)),
		Total: total,
	}
	for _, o := range list {
		cents, _ := o.AmountTotal.ToCents()
		amount := float64(cents) / 100
		resp.List = append(resp.List, v1.UserOrderListItem{
			OrderID:     o.ID,
			OrderNo:     o.OrderNo,
			ProductType: o.ProductType,
			Title:       o.TitleSnapshot,
			Amount:      amount,
			PaidAt:      formatOptionalTime(o.PaidAt),
			CreateAt:    formatTime(o.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}
