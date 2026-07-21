package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type JobHandler struct {
	*Handler
	jobService               service.JobService
	orderService             service.OrderService
	payService               service.PayService
	collectRepository        repository.CollectRepository
	contactHistoryRepository repository.ContactHistoryRepository
}

func NewJobHandler(
	handler *Handler,
	jobService service.JobService,
	orderService service.OrderService,
	payService service.PayService,
	collectRepository repository.CollectRepository,
	contactHistoryRepository repository.ContactHistoryRepository,
) *JobHandler {
	return &JobHandler{
		Handler:                  handler,
		jobService:               jobService,
		orderService:             orderService,
		payService:               payService,
		collectRepository:        collectRepository,
		contactHistoryRepository: contactHistoryRepository,
	}
}

// PrePublishRent godoc
// @Summary 发布招租（付费预下单）
// @Description 招租(biz_type=3)专用发布接口。一次事务预建 job(status=待支付) + rent_detail + order，返回微信支付参数。小程序完成支付后，微信回调将 job 翻转为 active 并刷新时间。photo_urls 至少1张、最多4张；transfer_fee_type=1 时 transfer_fee_amount 必填。发布价格从服务端配置 rent.publish_price 读取，客户端不传。
// @Tags 岗位模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.RentPrePublishRequest true "params"
// @Success 200 {object} v1.RentPrePublishResponse
// @Router /jobs/rent/pre_publish [post]
func (h *JobHandler) PrePublishRent(ctx *gin.Context) {
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
	var req v1.RentPrePublishRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := validatePhotoURLs(req.PhotoURLs); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	input := service.RentPrePublishInput{
		Positions:         req.Positions,
		Address:           req.Address,
		AddressDetail:     req.AddressDetail,
		Longitude:         req.Longitude,
		Latitude:          req.Latitude,
		Contact:           req.Contact,
		ContactPersonName: req.ContactPersonName,
		Description:       req.Description,
		PhotoURLs:         strings.Join(req.PhotoURLs, ","),
		FirstAreaID:       req.FirstAreaID,
		FirstAreaDes:      req.FirstAreaDes,
		SecondAreaID:      req.SecondAreaID,
		SecondAreaDes:     req.SecondAreaDes,
		ThirdAreaID:       req.ThirdAreaID,
		ThirdAreaDes:      req.ThirdAreaDes,
		FourAreaID:        req.FourAreaID,
		FourAreaDes:       req.FourAreaDes,
		MonthlyRent:       req.MonthlyRent,
		AreaSize:          req.AreaSize,
		TransferFeeType:   model.TransferFeeType(req.TransferFeeType),
		TransferFeeAmount: req.TransferFeeAmount,
		TransferDesc:      req.TransferDesc,
	}
	result, err := h.jobService.PrePublishRent(ctx, userID, openid, input)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.PrePublishRent error", zap.Error(err))
		if err == service.ErrInvalidRentInput {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	params, _ := result.PayParams.(v1.PayParams)
	v1.HandleSuccess(ctx, v1.RentPrePublishResponseData{
		JobID:     result.JobID,
		OrderID:   result.OrderID,
		OrderNo:   result.OrderNo,
		Amount:    result.Amount,
		PayParams: params,
	})
}

// Create godoc
// @Summary 发布岗位信息
// @Description 发布招聘或求职信息。biz_type: 1=招聘（默认）2=求职 3=招租，不传默认1。招聘时 company_name/address/longitude/latitude 必填；求职时留空即可。**招租(biz_type=3)必须走 /jobs/rent/pre_publish 付费流程，本接口不受理。** photo_urls 最多4张。每个用户招聘上限10条、求职上限5条（active状态）。
// @Tags 岗位模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobCreateRequest true "params"
// @Success 200 {object} v1.JobCreateResponse
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
	// biz_type 不传时默认为招聘
	if req.BizType == 0 {
		req.BizType = v1.BizTypeRecruit
	}
	// 招租禁止走通用 Create，必须走 /jobs/rent/pre_publish
	if req.BizType == v1.BizTypeRent {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "rent must be published via /jobs/rent/pre_publish")
		return
	}
	// 招聘时工作地点相关字段必填
	if req.BizType == v1.BizTypeRecruit {
		if req.CompanyName == "" || req.Address == "" || req.Longitude == 0 || req.Latitude == 0 {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "company_name, address, longitude, latitude are required for recruit")
			return
		}
	}
	input := service.JobCreateInput{
		BizType:            req.BizType,
		Positions:          req.Positions,
		CompanyName:        req.CompanyName,
		Longitude:          req.Longitude,
		Latitude:           req.Latitude,
		Address:            req.Address,
		AddressDetail:      req.AddressDetail,
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
		EnterpriseID:       req.EnterpriseID,
		RecruitNum:         req.RecruitNum,
		WorkContent:        req.WorkContent,
	}
	job, err := h.jobService.Create(ctx, userID, input)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.Create error", zap.Error(err))
		if err == service.ErrJobLimitExceeded {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		if err == service.ErrRentUseDedicatedAPI {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.JobCreateResponseData{
		JobID: job.ID,
	})
}

// Update godoc
// @Summary 修改岗位信息
// @Description 更新岗位字段，所有字段均为可选，传哪个改哪个。photo_urls 最多4张；basic_protection/salary_benefits/attendance_leave 传空数组表示清空。招租(biz_type=3)可传 rent_detail 更新月租、面积、转让费。
// @Tags 岗位模块
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
		ID:                req.ID,
		Positions:         req.Positions,
		CompanyName:       req.CompanyName,
		ContactPersonName: req.ContactPersonName,
		Longitude:         req.Longitude,
		Latitude:          req.Latitude,
		Address:           req.Address,
		AddressDetail:     req.AddressDetail,
		Contact:           req.Contact,
		Description:       req.Description,
		PhotoURLs:         nil,
		FirstAreaID:       req.FirstAreaID,
		FirstAreaDes:      req.FirstAreaDes,
		SecondAreaID:      req.SecondAreaID,
		SecondAreaDes:     req.SecondAreaDes,
		ThirdAreaID:       req.ThirdAreaID,
		ThirdAreaDes:      req.ThirdAreaDes,
		FourAreaID:        req.FourAreaID,
		FourAreaDes:       req.FourAreaDes,
		SalaryMin:         req.SalaryMin,
		SalaryMax:         req.SalaryMax,
		EnterpriseID:      req.EnterpriseID,
		RecruitNum:        req.RecruitNum,
		WorkContent:       req.WorkContent,
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
	if req.RentDetail != nil {
		input.RentDetail = &service.RentDetailUpdateInput{
			MonthlyRent:       req.RentDetail.MonthlyRent,
			AreaSize:          req.RentDetail.AreaSize,
			TransferFeeType:   model.TransferFeeType(req.RentDetail.TransferFeeType),
			TransferFeeAmount: req.RentDetail.TransferFeeAmount,
			TransferDesc:      req.RentDetail.TransferDesc,
		}
	}
	if err := h.jobService.Update(ctx, userID, input); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Update error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		if err == service.ErrInvalidRentInput {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

func validatePhotoURLs(urls []string) error {
	if len(urls) > 4 {
		return errors.New("photo_urls exceeds 4 images")
	}
	return nil
}

// Refresh godoc
// @Summary 免费刷新岗位
// @Description 将岗位的 refresh_time 更新为当前时间，信息流按刷新时间倒序，刷新后排名靠前。仅限岗位所有者操作。
// @Tags 岗位模块
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

// ShareRefresh godoc
// @Summary 分享刷新岗位
// @Description 每个自然日限使用1次，分享后免费刷新岗位排名（更新 refresh_time）。3001=今日次数已用完，403=非本人岗位。
// @Tags 岗位模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobShareRefreshRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/refresh/share [post]
func (h *JobHandler) ShareRefresh(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobShareRefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.jobService.ShareRefresh(ctx, userID, req.JobID); err != nil {
		switch err {
		case service.ErrShareRefreshLimitExceeded:
			v1.HandleError(ctx, http.StatusOK, v1.ErrShareRefreshLimitExceeded, nil)
		case service.ErrForbidden:
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
		default:
			h.logger.WithContext(ctx).Error("jobService.ShareRefresh error", zap.Error(err))
			v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		}
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// RefreshPay godoc
// @Description 创建付费刷新订单并返回微信支付参数，支付成功后后台回调自动完成刷新。price 单位：元。
// @Tags 岗位模块
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
	openid := GetOpenidFromCtx(ctx)
	if openid == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, "openid not found in token")
		return
	}

	var req v1.JobRefreshPayRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	// 招租(biz_type=3)不支持付费刷新
	if jobInfo, err := h.jobService.GetByID(ctx, req.JobID); err == nil && jobInfo != nil && jobInfo.BizType == v1.BizTypeRent {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "rent does not support paid refresh")
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

	// 获取金额（分）
	amountCents, err := order.AmountTotal.ToCents()
	if err != nil {
		h.logger.WithContext(ctx).Error("order.AmountTotal.ToCents error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}

	// 调用新的支付服务，获取支付参数
	params, err := h.payService.BuildPayParams(ctx, order.OrderNo, amountCents, openid, "付费刷新招聘")
	if err != nil {
		h.logger.WithContext(ctx).Error("payService.BuildPayParams error", zap.Error(err))
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
// @Summary 关闭岗位
// @Description 将岗位状态设为已关闭（status=2），记录关闭原因和关闭时间。关闭后可通过 /jobs/reopen 重新开启。仅限岗位所有者操作。
// @Tags 岗位模块
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
	if err := h.jobService.Close(ctx, userID, req.JobID, req.CloseReason); err != nil {
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

// Reopen godoc
// @Summary 重新开启岗位
// @Description 将已关闭的岗位状态重置为 active（status=1），清空关闭原因和关闭时间。仅限岗位所有者操作。
// @Tags 岗位模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobReopenRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/reopen [post]
func (h *JobHandler) Reopen(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobReopenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.jobService.Reopen(ctx, userID, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Reopen error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary 删除岗位
// @Description 将岗位状态设为已删除（status=4），软删除，不可恢复。仅限岗位所有者操作。
// @Tags 岗位模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobDeleteRequest true "params"
// @Success 200 {object} v1.Response
// @Router /jobs/delete [post]
func (h *JobHandler) Delete(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.JobDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.jobService.Delete(ctx, userID, req.JobID); err != nil {
		h.logger.WithContext(ctx).Error("jobService.Delete error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// GetCloseReasons godoc
// @Summary 获取关闭原因列表
// @Description 根据岗位类型返回对应的关闭原因枚举。type: 1=招聘 2=求职 3=招租，不传默认1。
// @Tags 通用接口
// @Accept json
// @Produce json
// @Param type query int true "岗位类型：1=招聘 2=求职 3=招租"
// @Success 200 {object} v1.JobCloseReasonResponse
// @Router /close_reasons [get]
func (h *JobHandler) GetCloseReasons(ctx *gin.Context) {
	bizTypeStr := ctx.Query("type")
	var bizType int
	if bizTypeStr == "" {
		bizType = v1.BizTypeRecruit // 默认招聘
	} else {
		fmt.Sscanf(bizTypeStr, "%d", &bizType)
	}

	var reasons []string
	switch bizType {
	case v1.BizTypeRecruit:
		reasons = []string{
			"招到人了",
			"不想招了",
			"效果不好，没人联系我",
			"信息有误，需要重新发布",
			"其他原因",
		}
	case v1.BizTypeResume:
		reasons = []string{
			"找到工作了",
			"暂时不想找工作了",
			"效果不好，没人联系我",
			"信息有误，需要重新发布",
			"其他原因",
		}
	case v1.BizTypeRent:
		reasons = []string{
			"租出去了",
			"暂时不租了",
			"效果不好，没人联系我",
			"信息有误，需要重新发布",
			"其他原因",
		}
	default:
		reasons = []string{}
	}

	v1.HandleSuccess(ctx, v1.JobCloseReasonResponseData{
		JobCloseReasonItem: v1.JobCloseReasonItem{
			Type:    bizType,
			Reasons: reasons,
		},
	})
}

// List godoc
// @Summary 岗位列表（带筛选+分页）
// @Description 公开接口，无需登录。filter.biz_type: 0=全部 1=招聘 2=求职 3=招租，不传默认全部。query_type: 1=置顶优先+刷新时间倒序 2=距离最近（需传 longitude/latitude） 3=最新发布，不传默认3。salary_min/salary_max 为0时不过滤薪资。basic_protection/salary_benefits/attendance_leave 数组，AND 过滤，多个值同时满足才返回。招租专属筛选(仅 biz_type=3 生效)：area_size_range(1=<15 2=[15,30) 3=[30,50) 4=[50,100) 5=[100,200) 6=>=200)，transfer_fee_flag(1=有 2=无)。
// @Tags 岗位模块
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
	salaryMax := max(req.Filter.SalaryMax, 0)
	query := repository.JobListQuery{
		BizType:         req.Filter.BizType,
		QueryType:       req.QueryType,
		Positions:       req.Filter.Positions,
		FirstAreaID:     req.Filter.FirstAreaID,
		SecondAreaID:    req.Filter.SecondAreaID,
		SalaryMin:       req.Filter.SalaryMin,
		SalaryMax:       salaryMax,
		BasicProtection: req.Filter.BasicProtection,
		SalaryBenefits:  req.Filter.SalaryBenefits,
		AttendanceLeave: req.Filter.AttendanceLeave,
		Longitude:       req.Filter.Longitude,
		Latitude:        req.Filter.Latitude,
		AreaSizeRange:   req.Filter.AreaSizeRange,
		TransferFeeFlag: req.Filter.TransferFeeFlag,
		PageNum:         req.PageNum,
		PageSize:        req.PageSize,
	}
	jobs, total, err := h.jobService.List(ctx, query)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.List error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	// 批量补齐招租扩展信息
	rentMap := h.batchLoadRentDetails(ctx, jobs)
	resp := v1.JobListResponseData{
		Jobs:  make([]v1.JobListItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		item := buildJobListItem(job)
		if job.BizType == v1.BizTypeRent {
			if d, ok := rentMap[job.ID]; ok && d != nil {
				dto := toRentDetailDTO(d)
				item.RentDetail = &dto
			}
		}
		resp.Jobs = append(resp.Jobs, item)
	}
	v1.HandleSuccess(ctx, resp)
}

// Info godoc
// @Summary 岗位详情
// @Description 返回岗位完整信息。已登录用户会额外返回 is_collected（0=未收藏 1=已收藏）和 is_contacted（0=未联系 1=已联系）。未登录时两者均为0。status=4（已删除）的岗位返回404。
// @Tags 岗位模块
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.JobInfoRequest true "params"
// @Success 200 {object} v1.JobListItem
// @Router /jobs/info [post]
func (h *JobHandler) Info(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)

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
	// 校验：删除状态的 job 不允许查询
	if job.Status == model.JobStatusDeleted {
		v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, "job not found")
		return
	}

	item := buildJobListItem(job)
	if job.BizType == v1.BizTypeRent {
		if d, err := h.jobService.GetRentDetailByJobID(ctx, job.ID); err == nil && d != nil {
			dto := toRentDetailDTO(d)
			item.RentDetail = &dto
		}
	}

	if userID > 0 {
		collect, err := h.collectRepository.Get(ctx, userID, job.ID, job.BizType)
		h.logger.WithContext(ctx).Info("collectRepository.Get", zap.Int64("userID", userID), zap.Int64("jobID", job.ID), zap.Any("collect", collect), zap.Error(err))
		if err == nil && collect != nil && collect.Status == model.CollectStatusActive {
			item.IsCollected = 1
		}
		contacted, err := h.contactHistoryRepository.ExistsByUserAndJob(ctx, userID, job.ID, job.BizType)
		h.logger.WithContext(ctx).Info("contactHistoryRepository.ExistsByUserAndJob", zap.Bool("contacted", contacted), zap.Error(err))
		if err == nil && contacted {
			item.IsContacted = 1
		}
	}

	v1.HandleSuccess(ctx, item)
}

// My godoc
// @Summary 我发布的岗位
// @Description 返回当前用户发布的岗位（不含已删除）。biz_type: 0=全部 1=招聘 2=求职，不传默认全部。status 数组：不传或空数组=全部非删除，传具体值按数组过滤。status 枚举：1=招聘中 2=已关闭 3=已禁用。
// @Tags 岗位模块
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
	jobs, total, err := h.jobService.ListByUser(ctx, userID, req.BizType, req.Status, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	rentMap := h.batchLoadRentDetails(ctx, jobs)
	resp := v1.JobMyResponseData{
		List:  make([]v1.JobMyItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		myItem := v1.JobMyItem{
			JobID:           job.ID,
			BizType:         job.BizType,
			Positions:       job.Positions,
			SalaryMin:       job.SalaryMin,
			SalaryMax:       job.SalaryMax,
			FirstAreaDes:    job.FirstAreaDes,
			SecondAreaDes:   job.SecondAreaDes,
			ThirdAreaDes:    job.ThirdAreaDes,
			Address:         job.Address,
			AddressDetail:   job.AddressDetail,
			CompanyName:     job.CompanyName,
			CreateAt:        formatTime(job.CreateAt),
			IsTop:           isJobTop(job),
			Status:          int(job.Status),
			LastRefreshTime: formatOptionalTime(job.RefreshTime),
			WorkContent:     job.WorkContent,
		}
		if job.BizType == v1.BizTypeRent {
			if d, ok := rentMap[job.ID]; ok && d != nil {
				dto := toRentDetailDTO(d)
				myItem.RentDetail = &dto
			}
		}
		resp.List = append(resp.List, myItem)
	}
	v1.HandleSuccess(ctx, resp)
}

// Top godoc
// @Summary 岗位置顶（付费）
// @Description 创建置顶订单并返回微信支付参数，支付成功后后台回调自动写入置顶时间窗口。top_hour 单位：小时，必须>0；price 单位：元，必须>0。仅限岗位所有者操作。
// @Tags 岗位模块
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
	openid := GetOpenidFromCtx(ctx)
	if openid == "" {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, "openid not found in token")
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
	// 招租(biz_type=3)不支持置顶
	if jobInfo, err := h.jobService.GetByID(ctx, req.JobID); err == nil && jobInfo != nil && jobInfo.BizType == v1.BizTypeRent {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "rent does not support top")
		return
	}
	order, _, err := h.orderService.CreateTopOrder(ctx, userID, req.JobID, req.TopHour, req.Price, req.ContactVoucherNum)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.CreateTopOrder error", zap.Error(err))
		if err == service.ErrForbidden {
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
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
	params, err := h.payService.BuildPayParams(ctx, order.OrderNo, amountCents, openid, "置顶招聘信息")
	if err != nil {
		h.logger.WithContext(ctx).Error("payService.BuildPayParams error", zap.Error(err))
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

// HomeTop godoc
// @Summary 首页置顶区
// @Description 公开接口，无需登录。返回当前有效置顶时间窗口内的岗位，按普通刷新和付费刷新中较晚的一次倒序；均未刷新时按发布时间排序。type: 0或不传=全部 1=招聘 2=求职，不传默认0。limit: 最大返回条数，不传默认5。
// @Tags 首页
// @Accept json
// @Produce json
// @Param request body v1.HomeTopRequest true "params"
// @Success 200 {object} v1.HomeTopResponseData
// @Router /home/top [post]
func (h *JobHandler) HomeTop(ctx *gin.Context) {
	var req v1.HomeTopRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	jobs, err := h.jobService.HomeTop(ctx, req.Type, req.FirstAreaID, req.SecondAreaID, req.Limit)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.HomeTop error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	rentMap := h.batchLoadRentDetails(ctx, jobs)
	resp := v1.HomeTopResponseData{
		List: make([]v1.JobListItem, 0, len(jobs)),
	}
	for _, job := range jobs {
		item := buildJobListItem(job)
		if job.BizType == v1.BizTypeRent {
			if d, ok := rentMap[job.ID]; ok && d != nil {
				dto := toRentDetailDTO(d)
				item.RentDetail = &dto
			}
		}
		resp.List = append(resp.List, item)
	}
	v1.HandleSuccess(ctx, resp)
}

// HomeFeed godoc
// @Summary 首页信息流
// @Description 公开接口，无需登录。返回非置顶的活跃岗位，按 COALESCE(refresh_time, create_at) 倒序排列，刷新过的优先。type: 0或不传=全部 1=招聘 2=求职，不传默认0。
// @Tags 首页
// @Accept json
// @Produce json
// @Param request body v1.HomeFeedRequest true "params"
// @Success 200 {object} v1.HomeFeedResponseData
// @Router /home/feed [post]
func (h *JobHandler) HomeFeed(ctx *gin.Context) {
	var req v1.HomeFeedRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	jobs, total, err := h.jobService.HomeFeed(ctx, req.Type, req.FirstAreaID, req.SecondAreaID, req.PageNum, req.PageSize)
	if err != nil {
		h.logger.WithContext(ctx).Error("jobService.HomeFeed error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	rentMap := h.batchLoadRentDetails(ctx, jobs)
	resp := v1.HomeFeedResponseData{
		List:  make([]v1.JobListItem, 0, len(jobs)),
		Total: total,
	}
	for _, job := range jobs {
		item := buildJobListItem(job)
		if job.BizType == v1.BizTypeRent {
			if d, ok := rentMap[job.ID]; ok && d != nil {
				dto := toRentDetailDTO(d)
				item.RentDetail = &dto
			}
		}
		resp.List = append(resp.List, item)
	}
	v1.HandleSuccess(ctx, resp)
}

func buildJobListItem(job *model.Job) v1.JobListItem {
	photos := splitCSV(job.PhotoURLs)
	basicProtection := splitCSV(job.BasicProtection)
	salaryBenefits := splitCSV(job.SalaryBenefits)
	attendanceLeave := splitCSV(job.AttendanceLeave)
	item := v1.JobListItem{
		ID:                job.ID,
		UserID:            job.UserID,
		BizType:           job.BizType,
		Positions:         job.Positions,
		CompanyName:       job.CompanyName,
		Longitude:         job.Longitude,
		Latitude:          job.Latitude,
		Address:           job.Address,
		AddressDetail:     job.AddressDetail,
		ContactPersonName: job.ContactPersonName,
		Contact:           job.Contact,
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
		BasicProtection:   basicProtection,
		SalaryBenefits:    salaryBenefits,
		AttendanceLeave:   attendanceLeave,
		Avatar:            job.Avatar,
		CreateAt:          formatTime(job.CreateAt),
		UpdateAt:          formatTime(job.UpdateAt),
		IsTop:             isJobTop(job),
		TopStartTime:      formatOptionalTime(job.TopStartTime),
		TopEndTime:        formatOptionalTime(job.TopEndTime),
		LastRefreshTime:   formatOptionalTime(job.RefreshTime),
		CloseReason:       job.CloseReason,
		CloseTime:         formatOptionalTime(job.CloseTime),
		EnterpriseID:      job.EnterpriseID,
		EnterpriseName:    job.EnterpriseName,
		RecruitNum:        job.RecruitNum,
		WorkContent:       job.WorkContent,
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

// batchLoadRentDetails 从 jobs 中筛出招租(biz_type=3)条目，批量读扩展表。
func (h *JobHandler) batchLoadRentDetails(ctx *gin.Context, jobs []*model.Job) map[int64]*model.RentDetail {
	ids := make([]int64, 0)
	for _, j := range jobs {
		if j.BizType == v1.BizTypeRent {
			ids = append(ids, j.ID)
		}
	}
	if len(ids) == 0 {
		return map[int64]*model.RentDetail{}
	}
	m, err := h.jobService.GetRentDetailsByJobIDs(ctx, ids)
	if err != nil {
		h.logger.WithContext(ctx).Warn("GetRentDetailsByJobIDs error", zap.Error(err))
		return map[int64]*model.RentDetail{}
	}
	return m
}

// toRentDetailDTO 将 model.RentDetail 转为 API DTO。
func toRentDetailDTO(d *model.RentDetail) v1.RentDetailDTO {
	return v1.RentDetailDTO{
		MonthlyRent:       d.MonthlyRent,
		AreaSize:          d.AreaSize,
		TransferFeeType:   int(d.TransferFeeType),
		TransferFeeAmount: d.TransferFeeAmount,
		TransferDesc:      d.TransferDesc,
	}
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

func maxTime(a, b *time.Time) *time.Time {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.After(*b) {
		return a
	}
	return b
}

func formatTimeMillis(ms int64) string {
	if ms <= 0 {
		return ""
	}
	return time.UnixMilli(ms).Format("2006-01-02 15:04:05.000")
}
