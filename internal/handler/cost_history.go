package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
)

type ContactVoucherHistoryHandler struct {
	*Handler
	contactVoucherHistoryService service.ContactVoucherHistoryService
	orderService                 service.OrderService
	contactHistoryService        service.ContactHistoryService
	payService                   service.PayService
}

func NewContactVoucherHistoryHandler(
	handler *Handler,
	contactVoucherHistoryService service.ContactVoucherHistoryService,
	orderService service.OrderService,
	contactHistoryService service.ContactHistoryService,
	payService service.PayService,
) *ContactVoucherHistoryHandler {
	return &ContactVoucherHistoryHandler{
		Handler:                      handler,
		contactVoucherHistoryService: contactVoucherHistoryService,
		orderService:                 orderService,
		contactHistoryService:        contactHistoryService,
		payService:                   payService,
	}
}

func (h *ContactVoucherHistoryHandler) GetContactVoucherHistory(ctx *gin.Context) {

}

// Buy godoc
// @Summary 联系券充值
// @Tags 联系券模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactVoucherBuyRequest true "params"
// @Success 200 {object} v1.PayOrderResponseData
// @Router /contact_voucher/buy [post]
func (h *ContactVoucherHistoryHandler) Buy(ctx *gin.Context) {
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
	params, err := h.payService.BuildJSAPIPayParams(ctx, order.OrderNo, req.Price)
	if err != nil {
		h.logger.WithContext(ctx).Error("payService.BuildJSAPIPayParams error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.PayOrderResponseData{
		OrderID:   order.ID,
		OrderNo:   order.OrderNo,
		Amount:    req.Price,
		PayParams: params,
	})
}

// Cost godoc
// @Summary 联系券消费
// @Tags 联系券模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactVoucherCostRequest true "params"
// @Success 200 {object} v1.Response
// @Router /contact_voucher/cost [post]
func (h *ContactVoucherHistoryHandler) Cost(ctx *gin.Context) {
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
	_, err := h.contactVoucherHistoryService.AdjustVoucher(ctx, userID, model.ContactVoucherHistoryCost, -1, "拨打电话")
	if err != nil {
		h.logger.WithContext(ctx).Error("contactVoucherHistoryService.AdjustVoucher error", zap.Error(err))
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

// Records godoc
// @Summary 我的券包
// @Tags 联系券模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactHistoryListRequest true "params"
// @Success 200 {object} v1.ContactVoucherRecordsResponseData
// @Router /contact_voucher/records [post]
func (h *ContactVoucherHistoryHandler) Records(ctx *gin.Context) {
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
	histories, total, err := h.contactVoucherHistoryService.ListByUser(ctx, userID, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("contactVoucherHistoryService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	num, err := h.contactVoucherHistoryService.GetUserVoucherNum(ctx, userID)
	if err != nil {
		h.logger.WithContext(ctx).Error("contactVoucherHistoryService.GetUserVoucherNum error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.ContactVoucherRecordsResponseData{
		ContactVoucherNum: num,
		List:              make([]v1.ContactVoucherRecordsItem, 0, len(histories)),
		ListTotal:         total,
	}
	for _, history := range histories {
		itemType := v1.ContactVoucherRecordCost
		title := "拨打电话"
		if history.BizType == model.ContactVoucherHistoryBuy {
			itemType = v1.ContactVoucherRecordBuy
			title = "购买"
		}
		resp.List = append(resp.List, v1.ContactVoucherRecordsItem{
			ID:        history.ID,
			Type:      itemType,
			Title:     title,
			ChangeNum: history.ChangeNum,
			CreateAt:  formatTime(history.CreateAt),
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
