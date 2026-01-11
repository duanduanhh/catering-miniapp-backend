package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
)

type CostHistoryHandler struct {
	*Handler
	costHistoryService service.CostHistoryService
	orderService       service.OrderService
	contactHistoryService service.ContactHistoryService
}

func NewCostHistoryHandler(
    handler *Handler,
    costHistoryService service.CostHistoryService,
    orderService service.OrderService,
    contactHistoryService service.ContactHistoryService,
) *CostHistoryHandler {
	return &CostHistoryHandler{
		Handler:      handler,
		costHistoryService: costHistoryService,
		orderService: orderService,
		contactHistoryService: contactHistoryService,
	}
}

func (h *CostHistoryHandler) GetCostHistory(ctx *gin.Context) {

}

func (h *CostHistoryHandler) Buy(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactVoucherBuyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	order, _, err := h.orderService.CreateContactVoucherOrder(ctx, userID, req.Price, req.ContactVoucherNum)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.CreateContactVoucherOrder error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.ContactVoucherBuyResponseData{
		OrderID:           order.ID,
		OrderNo:           order.OrderNo,
		BuyerUserID:       order.UserID,
		ContactVoucherNum: req.ContactVoucherNum,
		Price:             req.Price,
		CreatedAt:         formatTime(order.CreateAt),
	})
}

func (h *CostHistoryHandler) Cost(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactVoucherCostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	_, err := h.costHistoryService.AdjustVoucher(ctx, userID, service.CostHistoryConsume, -1, "拨打电话")
	if err != nil {
		h.logger.WithContext(ctx).Error("costHistoryService.AdjustVoucher error", zap.Error(err))
		if err == service.ErrInsufficientVoucher {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrInsufficientVoucher, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	if req.PurposeID != nil && req.PurposeType != nil {
		_, _ = h.contactHistoryService.Create(ctx, service.ContactHistoryCreateInput{
			UserID:           userID,
			PurposeID:        *req.PurposeID,
			PurposeType:      *req.PurposeType,
			PurposeUserID:    getInt64(req.PurposeUserID),
			PurposeUserPhone: getString(req.PurposeUserPhone),
		})
	}
	v1.HandleSuccess(ctx, nil)
}

func (h *CostHistoryHandler) My(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactHistoryListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	histories, total, err := h.costHistoryService.ListByUser(ctx, userID, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("costHistoryService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	num, err := h.costHistoryService.GetUserVoucherNum(ctx, userID)
	if err != nil {
		h.logger.WithContext(ctx).Error("costHistoryService.GetUserVoucherNum error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.ContactVoucherMyResponseData{
		ContactVoucherNum: num,
		List:              make([]v1.ContactVoucherMyItem, 0, len(histories)),
		ListTotal:         total,
	}
	for _, history := range histories {
		itemType := "consume"
		title := "拨打电话"
		if history.BizType == service.CostHistoryRecharge {
			itemType = "recharge"
			title = "购买"
		}
		resp.List = append(resp.List, v1.ContactVoucherMyItem{
			ID:       history.ID,
			Type:     itemType,
			Title:    title,
			ChangeNum: history.ChangeNum,
			CreateAt: formatTime(history.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

func getInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func getString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}
