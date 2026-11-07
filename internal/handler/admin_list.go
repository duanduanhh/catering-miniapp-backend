package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type AdminListHandler struct {
	*Handler
	adminListService service.AdminListService
}

func NewAdminListHandler(
	handler *Handler,
	adminListService service.AdminListService,
) *AdminListHandler {
	return &AdminListHandler{
		Handler:          handler,
		adminListService: adminListService,
	}
}

// reportReasonLabel 按业务类型把 reason 转成展示文案。
func reportReasonLabel(bizType, reason int) string {
	for _, r := range v1.ReportReasonsByBizType(bizType) {
		if r.Value == reason {
			return r.Label
		}
	}
	return ""
}

// contactFeedbackReasonLabel 把联系反馈 reason int 转成展示文案，按 bizType 区分文案
func contactFeedbackReasonLabel(bizType, reason int) string {
	cfg, ok := v1.ContactFeedbackConfigs[bizType]
	if !ok {
		return ""
	}
	for _, r := range cfg.Reasons {
		if r.Value == reason {
			return r.Label
		}
	}
	return ""
}

// ListUsers godoc
// @Summary  管理后台用户列表
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminUserListRequest true "params"
// @Success  200 {object} v1.AdminUserListResponseData
// @Router   /admin/users/list [post]
func (h *AdminListHandler) ListUsers(ctx *gin.Context) {
	var req v1.AdminUserListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	res, err := h.adminListService.Users(ctx, repository.AdminUserListQuery{
		UserID:    req.UserID,
		Keyword:   req.Keyword,
		Status:    req.Status,
		Type:      req.Type,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminListService.Users error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminUserListResponseData{List: make([]v1.AdminUserItem, 0, len(res.List)), Total: res.Total}
	for _, u := range res.List {
		resp.List = append(resp.List, v1.AdminUserItem{
			UserID:            u.ID,
			Avatar:            u.Avatar,
			Name:              u.Name,
			Phone:             u.Phone,
			UserCode:          u.UserCode,
			Sex:               u.Sex,
			Age:               u.Age,
			Type:              u.Type,
			Status:            u.Status,
			ContactVoucherNum: u.ContactVoucherNum,
			CollectNum:        u.CollectNum,
			BuyNum:            u.BuyNum,
			InviteNum:         u.InviteNum,
			InviterID:         u.InviterID,
			TotalRecharge:     u.TotalRecharge,
			Address:           u.Address,
			CreateAt:          formatTime(u.CreateAt),
			UpdateAt:          formatTime(u.UpdateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// ListOrders godoc
// @Summary  管理后台订单列表
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminOrderListRequest true "params"
// @Success  200 {object} v1.AdminOrderListResponseData
// @Router   /admin/orders/list [post]
func (h *AdminListHandler) ListOrders(ctx *gin.Context) {
	var req v1.AdminOrderListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	res, err := h.adminListService.Orders(ctx, repository.AdminOrderListQuery{
		OrderNo:       req.OrderNo,
		UserID:        req.UserID,
		UserKeyword:   req.UserKeyword,
		ProductType:   req.ProductType,
		Statuses:      req.Statuses,
		CreateAtStart: req.CreateAtStart,
		CreateAtEnd:   req.CreateAtEnd,
		PaidAtStart:   req.PaidAtStart,
		PaidAtEnd:     req.PaidAtEnd,
		PageNum:       req.PageNum,
		PageSize:      req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminListService.Orders error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminOrderListResponseData{List: make([]v1.AdminOrderItem, 0, len(res.List)), Total: res.Total}
	for _, order := range res.List {
		totalCents, _ := order.AmountTotal.ToCents()
		paidCents, _ := order.AmountPaid.ToCents()
		items := make([]v1.AdminOrderItemDetail, 0, len(res.Items[order.ID]))
		for _, item := range res.Items[order.ID] {
			items = append(items, v1.AdminOrderItemDetail{
				ID:                item.ID,
				ProductType:       int(item.ProductType),
				ProductID:         item.ProductID,
				SKUID:             item.SKUID,
				SKUCode:           item.SKUCode,
				SKUVersion:        item.SKUVersion,
				Title:             item.TitleSnapshot,
				UnitPrice:         item.UnitPriceSnapshot,
				TopHour:           item.TopHour,
				ContactVoucherNum: item.ContactVoucherNum,
				TargetType:        int(item.TargetType),
				TargetID:          item.TargetID,
			})
		}
		resp.List = append(resp.List, v1.AdminOrderItem{
			OrderID:     order.ID,
			OrderNo:     order.OrderNo,
			UserID:      order.UserID,
			UserName:    order.UserName,
			UserPhone:   order.UserPhone,
			AmountTotal: float64(totalCents) / 100,
			AmountPaid:  float64(paidCents) / 100,
			Currency:    order.Currency,
			Status:      int(order.Status),
			PayChannel:  order.PayChannel,
			PayTradeNo:  order.PayTradeNo,
			PaidAt:      formatOptionalTime(order.PaidAt),
			CanceledAt:  formatOptionalTime(order.CanceledAt),
			RefundedAt:  formatOptionalTime(order.RefundedAt),
			Remark:      order.Remark,
			CreateAt:    formatTime(order.CreateAt),
			Items:       items,
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// ListEnterprises godoc
// @Summary  管理后台企业列表
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminEnterpriseListRequest true "params"
// @Success  200 {object} v1.AdminEnterpriseListResponseData
// @Router   /admin/enterprises/list [post]
func (h *AdminListHandler) ListEnterprises(ctx *gin.Context) {
	var req v1.AdminEnterpriseListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	res, err := h.adminListService.Enterprises(ctx, repository.AdminEnterpriseListQuery{
		EnterpriseID: req.EnterpriseID,
		Keyword:      req.Keyword,
		Status:       req.Status,
		UserID:       req.UserID,
		PageNum:      req.PageNum,
		PageSize:     req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminListService.Enterprises error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminEnterpriseListResponseData{List: make([]v1.AdminEnterpriseItem, 0, len(res.List)), Total: res.Total}
	for _, e := range res.List {
		resp.List = append(resp.List, v1.AdminEnterpriseItem{
			ID:                  e.ID,
			UserID:              e.UserID,
			UserName:            e.UserName,
			UserPhone:           e.UserPhone,
			Name:                e.Name,
			SocialCreditCode:    e.SocialCreditCode,
			LegalRepresentative: e.LegalRepresentative,
			Address:             e.Address,
			BusinessScope:       e.BusinessScope,
			LicenseURL:          e.LicenseURL,
			IsDefault:           e.IsDefault,
			Status:              int(e.Status),
			CreateAt:            formatTime(e.CreateAt),
			UpdateAt:            formatTime(e.UpdateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// ListFeedbacks godoc
// @Summary  管理后台意见反馈列表
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminFeedbackListRequest true "params"
// @Success  200 {object} v1.AdminFeedbackListResponseData
// @Router   /admin/feedbacks/list [post]
func (h *AdminListHandler) ListFeedbacks(ctx *gin.Context) {
	var req v1.AdminFeedbackListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	res, err := h.adminListService.Feedbacks(ctx, repository.AdminFeedbackListQuery{
		FeedbackID: req.FeedbackID,
		Type:       req.Type,
		UserID:     req.UserID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminListService.Feedbacks error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminFeedbackListResponseData{List: make([]v1.AdminFeedbackItem, 0, len(res.List)), Total: res.Total}
	for _, f := range res.List {
		resp.List = append(resp.List, v1.AdminFeedbackItem{
			ID:        f.ID,
			UserID:    f.UserID,
			UserName:  f.UserName,
			UserPhone: f.UserPhone,
			Type:      int(f.Type),
			TypeName:  v1.FeedbackTypeName(f.Type),
			Content:   f.Content,
			PhotoURLs: splitCSV(f.PhotoURLs),
			CreateAt:  formatTime(f.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// ListContactHistories godoc
// @Summary  管理后台联系记录列表
// @Description 拨打电话记录列表，附带最新一条对应的拨打反馈聚合。
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminContactHistoryListRequest true "params"
// @Success  200 {object} v1.AdminContactHistoryListResponseData
// @Router   /admin/contact_histories/list [post]
func (h *AdminListHandler) ListContactHistories(ctx *gin.Context) {
	var req v1.AdminContactHistoryListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	res, err := h.adminListService.ContactHistories(ctx, repository.AdminContactHistoryListQuery{
		ID:            req.ID,
		UserID:        req.UserID,
		PurposeUserID: req.PurposeUserID,
		JobID:         req.JobID,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		PageNum:       req.PageNum,
		PageSize:      req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminListService.ContactHistories error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminContactHistoryListResponseData{List: make([]v1.AdminContactHistoryItem, 0, len(res.List)), Total: res.Total}
	for _, c := range res.List {
		fbCreate := ""
		if c.FeedbackCreateAt != nil {
			fbCreate = formatTime(*c.FeedbackCreateAt)
		}
		resp.List = append(resp.List, v1.AdminContactHistoryItem{
			ID:                  c.ID,
			UserID:              c.UserID,
			UserName:            c.UserName,
			UserPhone:           c.UserPhone,
			PurposeID:           c.PurposeID,
			PurposeType:         c.PurposeType,
			PurposeUserID:       c.PurposeUserID,
			PurposeUserName:     c.PurposeUserName,
			PurposeUserPhone:    c.PurposeUserPhone,
			UserDeleted:         c.UserDeleted,
			PurposeUserDeleted:  c.PurposeUserDeleted,
			CreateAt:            formatTime(c.CreateAt),
			FeedbackID:          c.FeedbackID,
			FeedbackReason:      c.FeedbackReason,
			FeedbackReasonLabel: contactFeedbackReasonLabel(c.PurposeType, c.FeedbackReason),
			FeedbackStatus:      c.FeedbackStatus,
			FeedbackCreateAt:    fbCreate,
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// ListReports godoc
// @Summary  管理后台举报列表
// @Tags     管理后台
// @Accept   json
// @Produce  json
// @Param    request body v1.AdminReportListRequest true "params"
// @Success  200 {object} v1.AdminReportListResponseData
// @Router   /admin/reports/list [post]
func (h *AdminListHandler) ListReports(ctx *gin.Context) {
	var req v1.AdminReportListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	res, err := h.adminListService.Reports(ctx, repository.AdminReportListQuery{
		ReportID:  req.ReportID,
		Status:    req.Status,
		Reason:    req.Reason,
		BizType:   req.BizType,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminListService.Reports error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminReportListResponseData{List: make([]v1.AdminReportItem, 0, len(res.List)), Total: res.Total}
	for _, r := range res.List {
		resp.List = append(resp.List, v1.AdminReportItem{
			ID:          r.ID,
			UserID:      r.UserID,
			UserName:    r.UserName,
			UserPhone:   r.UserPhone,
			JobID:       r.JobID,
			BizType:     r.BizType,
			Reason:      r.Reason,
			ReasonLabel: reportReasonLabel(r.BizType, r.Reason),
			Description: r.Description,
			Status:      int(r.Status),
			CreateAt:    formatTime(r.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}
