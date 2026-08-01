package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type ContactVoucherHistoryHandler struct {
	*Handler
	contactVoucherHistoryService service.ContactVoucherHistoryService
	orderService                 service.OrderService
	contactHistoryService        service.ContactHistoryService
	callbackHistoryService       service.CallbackHistoryService
	payService                   service.PayService
}

func NewContactVoucherHistoryHandler(
	handler *Handler,
	contactVoucherHistoryService service.ContactVoucherHistoryService,
	orderService service.OrderService,
	contactHistoryService service.ContactHistoryService,
	callbackHistoryService service.CallbackHistoryService,
	payService service.PayService,
) *ContactVoucherHistoryHandler {
	return &ContactVoucherHistoryHandler{
		Handler:                      handler,
		contactVoucherHistoryService: contactVoucherHistoryService,
		orderService:                 orderService,
		contactHistoryService:        contactHistoryService,
		callbackHistoryService:       callbackHistoryService,
		payService:                   payService,
	}
}

func (h *ContactVoucherHistoryHandler) GetContactVoucherHistory(ctx *gin.Context) {}

// Buy godoc
// @Summary 联系券充值
// @Description 传入套餐查询接口返回的 sku_code。价格、联系券数量和赠送数量均由服务端读取，支付成功后自动到账。
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
	openid := GetOpenidFromCtx(ctx)
	if openid == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, "openid not found in token")
		return
	}

	var req v1.ContactVoucherBuyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	order, _, err := h.orderService.CreateContactVoucherOrder(ctx, userID, req.SKUCode)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.CreateContactVoucherOrder error", zap.Error(err))
		if isPaymentPackageOrderError(err) {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		} else {
			v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		}
		return
	}

	// 获取金额（分）
	amountCents, err := order.AmountTotal.ToCents()
	if err != nil {
		h.logger.WithContext(ctx).Error("order.AmountTotal.ToCents error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	// 调用新的支付服务，获取支付参数
	params, err := h.payService.BuildPayParams(ctx, order.OrderNo, amountCents, openid, "购买联系券")
	if err != nil {
		h.logger.WithContext(ctx).Error("payService.BuildPayParams error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.PayOrderResponseData{
		OrderID:   order.ID,
		OrderNo:   order.OrderNo,
		Amount:    float64(amountCents) / 100,
		PayParams: params,
	})
}

// Cost godoc
// @Summary 联系券消费（拨打电话）
// @Description 消耗1张联系券，并可选记录联系历史。purpose_id/purpose_type/purpose_user_id/purpose_user_name/purpose_user_phone 均为可选，传入时自动写入联系记录。券余额不足时返回错误码 ErrInsufficientVoucher。
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
			PurposeUserName:  getString(req.PurposeUserName),
			PurposeUserPhone: getString(req.PurposeUserPhone),
		})
	}
	v1.HandleSuccess(ctx, nil)
}

// CallbackCost godoc
// @Summary 联系券消费（回拨电话）
// @Description 消耗1张联系券用于回拨电话，不写入联系记录表，仅记录到回拨记录表。券余额不足时返回错误码 ErrInsufficientVoucher。
// @Tags 联系券模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactVoucherCallbackCostRequest true "params"
// @Success 200 {object} v1.Response
// @Router /contact_voucher/callback_cost [post]
func (h *ContactVoucherHistoryHandler) CallbackCost(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactVoucherCallbackCostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	_, err := h.contactVoucherHistoryService.AdjustVoucher(ctx, userID, model.ContactVoucherHistoryCost, -1, "回拨")
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
		_ = h.callbackHistoryService.Create(ctx, service.CallbackHistoryCreateInput{
			UserID:           userID,
			PurposeID:        *req.PurposeID,
			PurposeType:      *req.PurposeType,
			PurposeUserID:    getInt64(req.PurposeUserID),
			PurposeUserName:  getString(req.PurposeUserName),
			PurposeUserPhone: getString(req.PurposeUserPhone),
		})
	}
	v1.HandleSuccess(ctx, nil)
}

// Records godoc
// @Summary 我的券包（联系券流水）
// @Description 返回当前用户的联系券余额和流水记录。响应中 type: 1=消费 2=充值；change_num 为变更张数（消费为负数，充值为正数）。
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
		if history.BizType == model.ContactVoucherHistoryBuy {
			itemType = v1.ContactVoucherRecordBuy
		}
		resp.List = append(resp.List, v1.ContactVoucherRecordsItem{
			ID:        history.ID,
			Type:      itemType,
			Title:     history.Remark,
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
