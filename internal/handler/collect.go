package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
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
		Handler:      handler,
		collectService: collectService,
	}
}

func (h *CollectHandler) GetCollect(ctx *gin.Context) {

}

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
	if err := h.collectService.Collect(ctx, userID, req.JobID, 1); err != nil {
		h.logger.WithContext(ctx).Error("collectService.Collect error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

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
	if err := h.collectService.Cancel(ctx, userID, req.JobID, 1); err != nil {
		h.logger.WithContext(ctx).Error("collectService.Cancel error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

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
		Jobs:  make([]v1.JobMyItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		resp.Jobs = append(resp.Jobs, v1.JobMyItem{
			ID:             job.ID,
			Positions:      job.Positions,
			SalaryMin:      job.SalaryMin,
			SalaryMax:      job.SalaryMax,
			FirstAreaDes:   job.FirstAreaDes,
			SecondAreaDes:  job.SecondAreaDes,
			ThirdAreaDes:   job.ThirdAreaDes,
			Address:        job.Address,
			CreateAt:       formatTime(job.CreateAt),
			IsTop:          job.IsTop,
			LastRefreshTime: formatTimeMillis(job.RefreshTime),
		})
	}
	v1.HandleSuccess(ctx, resp)
}
