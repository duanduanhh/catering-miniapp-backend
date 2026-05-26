package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type CollectHandler struct {
	*Handler
	collectService service.CollectService
}

func NewCollectHandler(
	handler *Handler,
	collectService service.CollectService,
) *CollectHandler {
	return &CollectHandler{
		Handler:        handler,
		collectService: collectService,
	}
}

func (h *CollectHandler) GetCollect(ctx *gin.Context) {}

// Collect godoc
// @Summary 收藏岗位
// @Description 收藏指定招聘岗位（仅支持招聘类型）。重复收藏同一岗位不会报错（幂等）。
// @Tags 收藏模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobCollectRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/collect [post]
func (h *CollectHandler) Collect(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobCollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.collectService.Collect(ctx, userID, req.JobID, req.BizType); err != nil {
		h.logger.WithContext(ctx).Error("collectService.Collect error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Cancel godoc
// @Summary 取消收藏
// @Description 取消收藏指定岗位。未收藏时取消不会报错（幂等）。
// @Tags 收藏模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobCancelCollectRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/cancel_collect [post]
func (h *CollectHandler) Cancel(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobCancelCollectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.collectService.Cancel(ctx, userID, req.JobID, req.BizType); err != nil {
		h.logger.WithContext(ctx).Error("collectService.Cancel error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// My godoc
// @Summary 我收藏的岗位
// @Description 返回当前用户的收藏列表。biz_type: 0=全部 1=招聘（当前仅支持招聘），不传默认全部。
// @Tags 收藏模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.CollectMyRequest true "params"
// @Success 200 {object} v1.CollectMyResponseData
// @Router /collect/my [post]
func (h *CollectHandler) My(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.CollectMyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	jobs, total, err := h.collectService.ListByUser(ctx, userID, req.BizType, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("collectService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.CollectMyResponseData{
		List:  make([]v1.JobMyItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		resp.List = append(resp.List, v1.JobMyItem{
			JobID:             job.ID,
			BizType:           job.BizType,
			Positions:         job.Positions,
			SalaryMin:         job.SalaryMin,
			SalaryMax:         job.SalaryMax,
			FirstAreaDes:      job.FirstAreaDes,
			SecondAreaDes:     job.SecondAreaDes,
			ThirdAreaDes:      job.ThirdAreaDes,
			Address:           job.Address,
			CreateAt:          formatTime(job.CreateAt),
			IsTop:             isJobTop(job),
			Status:            int(job.Status),
			LastRefreshTime:   formatOptionalTime(job.RefreshTime),
			ContactPersonName: job.ContactPersonName,
			Contact:           job.Contact,
			Avatar:            job.Avatar,
			BasicProtection:   splitCSV(job.BasicProtection),
			SalaryBenefits:    splitCSV(job.SalaryBenefits),
			AttendanceLeave:   splitCSV(job.AttendanceLeave),
			UserID:            job.UserID,
			Description:       job.Description,
		})
	}
	v1.HandleSuccess(ctx, resp)
}
