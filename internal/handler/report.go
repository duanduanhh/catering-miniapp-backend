package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type ReportHandler struct {
	*Handler
	reportService service.ReportService
}

func NewReportHandler(handler *Handler, reportService service.ReportService) *ReportHandler {
	return &ReportHandler{
		Handler:       handler,
		reportService: reportService,
	}
}

// Reasons godoc
// @Summary 获取举报原因选项列表
// @Description 返回举报表单的原因选项列表；biz_type=3 时返回招租专用原因，未传时兼容返回岗位举报原因
// @Tags 举报
// @Produce json
// @Success 200 {object} v1.ReportReasonsResponseData
// @Param biz_type query int false "业务类型：1=招聘 2=求职 3=招租"
// @Router /report/reasons [get]
func (h *ReportHandler) Reasons(ctx *gin.Context) {
	var req v1.ReportReasonsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.ReportReasonsResponseData{
		Reasons: v1.ReportReasonsByBizType(req.BizType),
	})
}

// Submit godoc
// @Summary 提交举报
// @Description 对岗位进行违规举报。reason 值域：1=诈骗/收费 2=薪资虚假 3=代招冒充直招 4=无法联系到招聘方 5=其他
// @Tags 举报
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.ReportSubmitRequest true "params"
// @Success 200 {object} v1.Response
// @Router /report/submit [post]
func (h *ReportHandler) Submit(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.ReportSubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.ReportSubmitInput{
		JobID:       req.JobID,
		BizType:     req.BizType,
		Reason:      req.Reason,
		Description: req.Description,
	}
	if err := h.reportService.Submit(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("reportService.Submit error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}
