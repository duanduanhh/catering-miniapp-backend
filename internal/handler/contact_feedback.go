package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type ContactFeedbackHandler struct {
	*Handler
	contactFeedbackService service.ContactFeedbackService
}

func NewContactFeedbackHandler(handler *Handler, contactFeedbackService service.ContactFeedbackService) *ContactFeedbackHandler {
	return &ContactFeedbackHandler{
		Handler:                handler,
		contactFeedbackService: contactFeedbackService,
	}
}

// Reasons godoc
// @Summary 获取信息反馈原因选项列表
// @Description 根据 biz_type 返回对应的反馈原因选项。biz_type: 1=招聘 2=求职 3=招租
// @Tags 信息反馈
// @Produce json
// @Param biz_type query int true "业务类型: 1=招聘 2=求职 3=招租"
// @Success 200 {object} v1.ContactFeedbackConfig
// @Router /contact_feedback/reasons [get]
func (h *ContactFeedbackHandler) Reasons(ctx *gin.Context) {
	bizTypeStr := ctx.Query("biz_type")
	bizType, _ := strconv.Atoi(bizTypeStr)
	config, ok := v1.ContactFeedbackConfigs[bizType]
	if !ok {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "invalid biz_type")
		return
	}
	v1.HandleSuccess(ctx, config)
}

// Submit godoc
// @Summary 提交信息反馈
// @Description 联系对方后提交信息反馈。biz_type: 1=招聘 2=求职 3=招租。reason: 1=电话无法接通 2=空号/错误号码 3=对方表示已不再需要 4=信息严重不符 5=其他
// @Tags 信息反馈
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ContactFeedbackSubmitRequest true "params"
// @Success 200 {object} v1.Response
// @Router /contact_feedback/submit [post]
func (h *ContactFeedbackHandler) Submit(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ContactFeedbackSubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.ContactFeedbackSubmitInput{
		JobID:       req.JobID,
		BizType:     req.BizType,
		Reason:      req.Reason,
		Description: req.Description,
	}
	if err := h.contactFeedbackService.Submit(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("contactFeedbackService.Submit error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}
