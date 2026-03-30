package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type ContactHistoryHandler struct {
	*Handler
	contactHistoryService service.ContactHistoryService
}

func NewContactHistoryHandler(
	handler *Handler,
	contactHistoryService service.ContactHistoryService,
) *ContactHistoryHandler {
	return &ContactHistoryHandler{
		Handler:               handler,
		contactHistoryService: contactHistoryService,
	}
}

func (h *ContactHistoryHandler) GetContactHistory(ctx *gin.Context) {}

// ListOut godoc
// @Summary 我联系的
// @Tags 联系模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactHistoryListRequest true "params"
// @Success 200 {object} v1.ContactHistoryListResponseData
// @Router /contact_history/out [post]
func (h *ContactHistoryHandler) ListOut(ctx *gin.Context) {
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
	items, total, err := h.contactHistoryService.ListOut(ctx, userID, req.BizType, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("contactHistoryService.ListOut error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.ContactHistoryListResponseData{
		List:  make([]v1.ContactHistoryItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		resp.List = append(resp.List, v1.ContactHistoryItem{
			ID:               item.ID,
			Positions:        item.Positions,
			Address:          item.Address,
			PurposeUserID:    item.PurposeUserID,
			PurposeUserName:  item.PurposeUserName,
			PurposeUserPhone: item.PurposeUserPhone,
			CreateAt:         formatTime(item.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// ListIn godoc
// @Summary 联系我的
// @Tags 联系模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactHistoryListRequest true "params"
// @Success 200 {object} v1.ContactHistoryListResponseData
// @Router /contact_history/in [post]
func (h *ContactHistoryHandler) ListIn(ctx *gin.Context) {
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
	items, total, err := h.contactHistoryService.ListIn(ctx, userID, req.BizType, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("contactHistoryService.ListIn error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.ContactHistoryListResponseData{
		List:  make([]v1.ContactHistoryItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		resp.List = append(resp.List, v1.ContactHistoryItem{
			ID:              item.ID,
			Positions:       item.Positions,
			Address:         item.Address,
			PurposeUserName: item.PurposeUserName,
			CreateAt:        formatTime(item.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// DeleteOut godoc
// @Summary 删除我联系的
// @Tags 联系模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactHistoryDeleteRequest true "params"
// @Success 200 {object} v1.Response
// @Router /contact_history/out [delete]
func (h *ContactHistoryHandler) DeleteOut(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactHistoryDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.contactHistoryService.DeleteOut(ctx, userID, req.PurposeID); err != nil {
		h.logger.WithContext(ctx).Error("contactHistoryService.DeleteOut error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// DeleteIn godoc
// @Summary 删除联系我的
// @Tags 联系模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactHistoryDeleteRequest true "params"
// @Success 200 {object} v1.Response
// @Router /contact_history/in [delete]
func (h *ContactHistoryHandler) DeleteIn(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactHistoryDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.contactHistoryService.DeleteIn(ctx, userID, req.PurposeID); err != nil {
		h.logger.WithContext(ctx).Error("contactHistoryService.DeleteIn error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}
