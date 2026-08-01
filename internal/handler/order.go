package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type OrderHandler struct {
	*Handler
	orderService service.OrderService
}

func NewOrderHandler(
	handler *Handler,
	orderService service.OrderService,
) *OrderHandler {
	return &OrderHandler{
		Handler:      handler,
		orderService: orderService,
	}
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
