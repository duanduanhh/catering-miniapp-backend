package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"go.uber.org/zap"
)

type JobHandler struct {
	*Handler
	jobService   service.JobService
	orderService service.OrderService
	payService   service.PayService
}

func NewJobHandler(
	handler *Handler,
	jobService service.JobService,
	orderService service.OrderService,
	payService service.PayService,
) *JobHandler {
	return &JobHandler{
		Handler:      handler,
		jobService:   jobService,
		orderService: orderService,
		payService:   payService,
	}
}

// Create godoc
// @Summary 发布招聘信息
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobCreateRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/create [post]
func (h *JobHandler) Create(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := validatePhotoURLs(req.PhotoURLs); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.JobCreateInput{
		Positions:          req.Positions,
		CompanyName:        req.CompanyName,
		Longitude:          req.Longitude,
		Latitude:           req.Latitude,
		Address:            req.Address,
		Contact:            req.Contact,
		ContanctPersonName: req.ContactPersonName,
		Description:        req.Description,
		PhotoURLs:          strings.Join(req.PhotoURLs, ","),
		FirstAreaID:        req.FirstAreaID,
		FirstAreaDes:       req.FirstAreaDes,
		SecondAreaID:       req.SecondAreaID,
		SecondAreaDes:      req.SecondAreaDes,
		ThirdAreaID:        req.ThirdAreaID,
		ThirdAreaDes:       req.ThirdAreaDes,
		FourAreaID:         req.FourAreaID,
		FourAreaDes:        req.FourAreaDes,
		SalaryMin:          req.SalaryMin,
		SalaryMax:          req.SalaryMax,
		BasicProtection:    strings.Join(req.BasicProtection, ","),
		SalaryBenefits:     strings.Join(req.SalaryBenefits, ","),
		AttendanceLeave:    strings.Join(req.AttendanceLeave, ","),
	}
	if _, err := h.jobService.Create(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Create error", zap.Error(err))
		if err == service.ErrJobLimitExceeded {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Update godoc
// @Summary 修改招聘信息
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobUpdateRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/update [post]
func (h *JobHandler) Update(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if req.PhotoURLs != nil {
		if err := validatePhotoURLs(req.PhotoURLs); err != nil {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
	}
	input := service.JobUpdateInput{
		ID:            req.ID,
		Positions:     req.Positions,
		Longitude:     req.Longitude,
		Latitude:      req.Latitude,
		Address:       req.Address,
		Contact:       req.Contact,
		Description:   req.Description,
		PhotoURLs:     nil,
		FirstAreaID:   req.FirstAreaID,
		FirstAreaDes:  req.FirstAreaDes,
		SecondAreaID:  req.SecondAreaID,
		SecondAreaDes: req.SecondAreaDes,
		ThirdAreaID:   req.ThirdAreaID,
		ThirdAreaDes:  req.ThirdAreaDes,
		FourAreaID:    req.FourAreaID,
		FourAreaDes:   req.FourAreaDes,
		SalaryMin:     req.SalaryMin,
		SalaryMax:     req.SalaryMax,
	}
	if req.PhotoURLs != nil {
		joined := strings.Join(req.PhotoURLs, ",")
		input.PhotoURLs = &joined
	}
	if req.BasicProtection != nil {
		joined := strings.Join(req.BasicProtection, ",")
		input.BasicProtection = &joined
	}
	if req.SalaryBenefits != nil {
		joined := strings.Join(req.SalaryBenefits, ",")
		input.SalaryBenefits = &joined
	}
	if req.AttendanceLeave != nil {
		joined := strings.Join(req.AttendanceLeave, ",")
		input.AttendanceLeave = &joined
	}
	if err := h.jobService.Update(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Update error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

func validatePhotoURLs(urls []string) error {
	if len(urls) > 4 {
		return fmt.Errorf("photo_urls exceeds 4 images")
	}
	return nil
}

// Refresh godoc
// @Summary 刷新招聘信息
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobRefreshRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/refresh [post]
func (h *JobHandler) Refresh(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobRefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.jobService.Refresh(ctx, userID, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Refresh error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// RefreshPay godoc
// @Summary 付费刷新招聘信息
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobRefreshPayRequest true "params"
// @Success 200 {object} v1.PayOrderResponseData
// @Router /jobs/refresh/pay [post]
func (h *JobHandler) RefreshPay(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobRefreshPayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	order, _, err := h.orderService.CreateRefreshOrder(ctx, userID, req.JobID, req.Price)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.CreateRefreshOrder error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
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

// Close godoc
// @Summary 关闭招聘信息
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobCloseRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/close [post]
func (h *JobHandler) Close(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobCloseRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.jobService.Close(ctx, userID, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Close error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// List godoc
// @Summary 招聘信息列表
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Param request body v1.JobListRequest true "params"
// @Success 200 {object} v1.JobListResponse
// @Router /jobs/list [post]
func (h *JobHandler) List(ctx *gin.Context) {
	var req v1.JobListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	salaryMax := req.Filter.SalaryMax
	if salaryMax < 0 {
		salaryMax = 0
	}
	query := repository.JobListQuery{
		QueryType:       req.QueryType,
		Positions:       req.Filter.Positions,
		City:            req.Filter.City,
		SalaryMin:       req.Filter.SalaryMin,
		SalaryMax:       salaryMax,
		BasicProtection: req.Filter.BasicProtection,
		SalaryBenefits:  req.Filter.SalaryBenefits,
		AttendanceLeave: req.Filter.AttendanceLeave,
		Longitude:       req.Filter.Longitude,
		Latitude:        req.Filter.Latitude,
		PageNum:         req.PageNum,
		PageSize:        req.PageSize,
	}
	jobs, total, err := h.jobService.List(ctx, query)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.List error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.JobListResponseData{
		Jobs:  make([]v1.JobListItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		resp.Jobs = append(resp.Jobs, buildJobListItem(job))
	}
	v1.HandleSuccess(ctx, resp)
}

// Info godoc
// @Summary 招聘信息详情
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Param request body v1.JobInfoRequest true "params"
// @Success 200 {object} v1.JobListItem
// @Router /jobs/info [post]
func (h *JobHandler) Info(ctx *gin.Context) {
	var req v1.JobInfoRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	job, err := h.jobService.GetByID(ctx, req.JobID)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.GetByID error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	item := buildJobListItem(job)
	v1.HandleSuccess(ctx, item)
}

// My godoc
// @Summary 我发布的
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobMyRequest true "params"
// @Success 200 {object} v1.JobMyResponseData
// @Router /jobs/my [post]
func (h *JobHandler) My(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobMyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	jobs, total, err := h.jobService.ListByUser(ctx, userID, req.BizType, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	resp := v1.JobMyResponseData{
		List:  make([]v1.JobMyItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		resp.List = append(resp.List, v1.JobMyItem{
			JobID:           job.ID,
			Positions:       job.Positions,
			SalaryMin:       job.SalaryMin,
			SalaryMax:       job.SalaryMax,
			FirstAreaDes:    job.FirstAreaDes,
			SecondAreaDes:   job.SecondAreaDes,
			ThirdAreaDes:    job.ThirdAreaDes,
			Address:         job.Address,
			CreateAt:        formatTime(job.CreateAt),
			IsTop:           isJobTop(job),
			LastRefreshTime: formatOptionalTime(job.RefreshTime),
		})
	}
	v1.HandleSuccess(ctx, resp)
}

// Top godoc
// @Summary 置顶招聘信息
// @Tags 招聘模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobTopRequest true "params"
// @Success 200 {object} v1.PayOrderResponseData
// @Router /jobs/top [post]
func (h *JobHandler) Top(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobTopRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if req.TopHour <= 0 || req.Price <= 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "top_hour and price must be positive")
		return
	}
	order, _, err := h.orderService.CreateTopOrder(ctx, userID, req.JobID, req.TopHour, req.Price)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.CreateTopOrder error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
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

func buildJobListItem(job *model.Job) v1.JobListItem {
	photos := splitCSV(job.PhotoURLs)
	item := v1.JobListItem{
		ID:                job.ID,
		UserID:            job.UserID,
		Positions:         job.Positions,
		Longitude:         job.Longitude,
		Latitude:          job.Latitude,
		Address:           job.Address,
		Contact:           job.Contact,
		ContactPersonName: job.ContactPersonName,
		Description:       job.Description,
		PhotoURLs:         photos,
		Status:            job.Status,
		FirstAreaID:       job.FirstAreaID,
		FirstAreaDes:      job.FirstAreaDes,
		SecondAreaID:      job.SecondAreaID,
		SecondAreaDes:     job.SecondAreaDes,
		ThirdAreaID:       job.ThirdAreaID,
		ThirdAreaDes:      job.ThirdAreaDes,
		FourAreaID:        job.FourAreaID,
		FourAreaDes:       job.FourAreaDes,
		SalaryMin:         job.SalaryMin,
		SalaryMax:         job.SalaryMax,
		CreateAt:          formatTime(job.CreateAt),
		UpdateAt:          formatTime(job.UpdateAt),
		IsTop:             isJobTop(job),
		TopStartTime:      formatOptionalTime(job.TopStartTime),
		TopEndTime:        formatOptionalTime(job.TopEndTime),
		LastRefreshTime:   formatOptionalTime(job.RefreshTime),
	}
	return item
}

func isJobTop(job *model.Job) int {
	if job == nil || job.TopStartTime == nil || job.TopEndTime == nil {
		return 0
	}
	now := time.Now()
	if now.Before(*job.TopStartTime) || now.After(*job.TopEndTime) {
		return 0
	}
	return 1
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, item := range parts {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + "****" + phone[len(phone)-4:]
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05.000")
}

func formatOptionalTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05.000")
}

func formatTimeMillis(ms int64) string {
	if ms <= 0 {
		return ""
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04:05.000")
}
