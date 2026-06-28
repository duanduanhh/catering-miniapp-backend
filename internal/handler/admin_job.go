package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type AdminJobHandler struct {
	*Handler
	adminJobService service.AdminJobService
}

func NewAdminJobHandler(
	handler *Handler,
	adminJobService service.AdminJobService,
) *AdminJobHandler {
	return &AdminJobHandler{
		Handler:         handler,
		adminJobService: adminJobService,
	}
}

// AdminListJobs godoc
// @Summary 后台岗位列表
// @Description 管理后台查询招聘/求职岗位列表，支持按类型、状态、关键词筛选。
// @Tags 管理后台
// @Accept json
// @Produce json
// @Param request body v1.AdminJobListRequest true "params"
// @Success 200 {object} v1.AdminJobListResponseData
// @Router /admin/jobs/list [post]
func (h *AdminJobHandler) List(ctx *gin.Context) {
	var req v1.AdminJobListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	items, total, err := h.adminJobService.List(ctx, repository.AdminJobListQuery{
		JobID:    req.JobID,
		UserID:   req.UserID,
		BizType:  req.BizType,
		Status:   req.Status,
		Keyword:  req.Keyword,
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("adminJobService.List error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.AdminJobListResponseData{
		List:  make([]v1.AdminJobItem, 0, len(items)),
		Total: total,
	}
	for _, item := range items {
		resp.List = append(resp.List, v1.AdminJobItem{
			JobID:         item.JobID,
			BizType:       item.BizType,
			Positions:     item.Positions,
			CompanyName:   item.CompanyName,
			Address:       item.Address,
			SalaryMin:     item.SalaryMin,
			SalaryMax:     item.SalaryMax,
			Status:        item.Status,
			UserID:        item.UserID,
			UserName:      item.UserName,
			UserPhone:     item.UserPhone,
			CreateAt:      formatTime(item.CreateAt),
			UpdateAt:      formatTime(item.UpdateAt),
			FirstAreaDes:  item.FirstAreaDes,
			SecondAreaDes: item.SecondAreaDes,
			ThirdAreaDes:  item.ThirdAreaDes,
			Description:   item.Description,
			PhotoURLs:     splitCSV(item.PhotoURLs),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// AdminDisableJob godoc
// @Summary 禁用岗位
// @Tags 管理后台
// @Accept json
// @Produce json
// @Param request body v1.AdminJobIDRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/jobs/disable [post]
func (h *AdminJobHandler) Disable(ctx *gin.Context) {
	var req v1.AdminJobIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.adminJobService.Disable(ctx, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("adminJobService.Disable error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminEnableJob godoc
// @Summary 恢复岗位
// @Tags 管理后台
// @Accept json
// @Produce json
// @Param request body v1.AdminJobIDRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/jobs/enable [post]
func (h *AdminJobHandler) Enable(ctx *gin.Context) {
	var req v1.AdminJobIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.adminJobService.Enable(ctx, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("adminJobService.Enable error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminDeleteJob godoc
// @Summary 删除岗位
// @Tags 管理后台
// @Accept json
// @Produce json
// @Param request body v1.AdminJobIDRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/jobs/delete [post]
func (h *AdminJobHandler) Delete(ctx *gin.Context) {
	var req v1.AdminJobIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.adminJobService.Delete(ctx, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("adminJobService.Delete error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}
