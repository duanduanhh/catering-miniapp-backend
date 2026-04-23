package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type FeedbackHandler struct {
	*Handler
	feedbackService service.FeedbackService
}

func NewFeedbackHandler(handler *Handler, feedbackService service.FeedbackService) *FeedbackHandler {
	return &FeedbackHandler{
		Handler:         handler,
		feedbackService: feedbackService,
	}
}

// Submit godoc
// @Summary 提交意见反馈
// @Description type 值域：1=产品建议 2=功能问题 3=内容修正 4=其他
// @Tags 意见反馈
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.FeedbackSubmitRequest true "params"
// @Success 200 {object} v1.Response
// @Router /feedback/submit [post]
func (h *FeedbackHandler) Submit(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.FeedbackSubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := validatePhotoURLs(req.PhotoURLs); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.FeedbackSubmitInput{
		Type:      model.FeedbackType(req.Type),
		Content:   req.Content,
		PhotoURLs: req.PhotoURLs,
	}
	if err := h.feedbackService.Submit(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("feedbackService.Submit error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// List godoc
// @Summary 我的意见反馈列表
// @Tags 意见反馈
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.FeedbackListRequest true "params"
// @Success 200 {object} v1.FeedbackListResponseData
// @Router /feedback/my [post]
func (h *FeedbackHandler) List(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.FeedbackListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	list, total, err := h.feedbackService.ListByUser(ctx, userID, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("feedbackService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.FeedbackListResponseData{
		List:  make([]v1.FeedbackListItem, 0, len(list)),
		Total: total,
	}
	for _, f := range list {
		resp.List = append(resp.List, v1.FeedbackListItem{
			ID:        f.ID,
			Type:      int(f.Type),
			TypeName:  v1.FeedbackTypeName(f.Type),
			Content:   f.Content,
			PhotoURLs: splitCSV(f.PhotoURLs),
			CreateAt:  formatTime(f.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, resp)
}
