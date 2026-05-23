package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type EnterpriseHandler struct {
	*Handler
	enterpriseService service.EnterpriseService
}

func NewEnterpriseHandler(handler *Handler, enterpriseService service.EnterpriseService) *EnterpriseHandler {
	return &EnterpriseHandler{
		Handler:           handler,
		enterpriseService: enterpriseService,
	}
}

// OCR godoc
// @Summary 营业执照 OCR 识别
// @Description 传入已上传到 OSS 的营业执照图片 URL，调用阿里云 OCR 识别企业工商信息并返回结构化字段，供前端回填表单（用户可二次修改）。识别失败时返回空字段，不阻断流程。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.EnterpriseOCRRequest true "params"
// @Success 200 {object} v1.EnterpriseOCRResponseData
// @Router /enterprise/ocr [post]
func (h *EnterpriseHandler) OCR(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.EnterpriseOCRRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	result, err := h.enterpriseService.OCR(ctx, req.LicenseURL)
	if err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.OCR error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrEnterpriseOCRFailed, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.EnterpriseOCRResponseData{
		Name:                result.Name,
		SocialCreditCode:    result.SocialCreditCode,
		LegalRepresentative: result.LegalRepresentative,
		Address:             result.Address,
		EstablishedDate:     result.EstablishedDate,
		BusinessPeriod:      result.BusinessPeriod,
		RegisteredCapital:   result.RegisteredCapital,
		BusinessScope:       result.BusinessScope,
	})
}

// Create godoc
// @Summary 创建企业（提交认证）
// @Description 提交营业执照信息创建企业，social_credit_code 为18位统一社会信用代码，同一用户不可重复添加相同代码。is_default=1 时自动清除其他企业的默认状态。当前版本直接认证通过（status=2）。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.EnterpriseCreateRequest true "params"
// @Success 200 {object} v1.EnterpriseCreateResponseData
// @Router /enterprise/create [post]
func (h *EnterpriseHandler) Create(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.EnterpriseCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	e, err := h.enterpriseService.Create(ctx, userID, service.EnterpriseCreateInput{
		Name:                req.Name,
		SocialCreditCode:    req.SocialCreditCode,
		LegalRepresentative: req.LegalRepresentative,
		Address:             req.Address,
		EstablishedDate:     req.EstablishedDate,
		BusinessPeriod:      req.BusinessPeriod,
		RegisteredCapital:   req.RegisteredCapital,
		BusinessScope:       req.BusinessScope,
		LicenseURL:          req.LicenseURL,
		IsDefault:           req.IsDefault,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.Create error", zap.Error(err))
		switch err {
		case service.ErrInvalidCreditCode:
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrInvalidCreditCode, err.Error())
		case service.ErrEnterpriseDuplicate:
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrEnterpriseDuplicate, err.Error())
		default:
			v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		}
		return
	}
	v1.HandleSuccess(ctx, v1.EnterpriseCreateResponseData{EnterpriseID: e.ID})
}

// Update godoc
// @Summary 修改企业信息
// @Description 更新企业字段，所有字段均为可选，传哪个改哪个。social_credit_code 不允许修改。is_default=1 时自动清除其他企业的默认状态。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.EnterpriseUpdateRequest true "params"
// @Success 200 {object} v1.Response
// @Router /enterprise/update [post]
func (h *EnterpriseHandler) Update(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.EnterpriseUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	err := h.enterpriseService.Update(ctx, userID, service.EnterpriseUpdateInput{
		ID:                  req.ID,
		Name:                req.Name,
		LegalRepresentative: req.LegalRepresentative,
		Address:             req.Address,
		EstablishedDate:     req.EstablishedDate,
		BusinessPeriod:      req.BusinessPeriod,
		RegisteredCapital:   req.RegisteredCapital,
		BusinessScope:       req.BusinessScope,
		LicenseURL:          req.LicenseURL,
		IsDefault:           req.IsDefault,
	})
	if err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.Update error", zap.Error(err))
		switch err {
		case service.ErrForbidden:
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
		case service.ErrEnterpriseNotFound:
			v1.HandleError(ctx, http.StatusNotFound, v1.ErrEnterpriseNotFound, err.Error())
		default:
			v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		}
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// Delete godoc
// @Summary 删除企业
// @Description 软删除企业（status=4），若被删除的是默认企业则同时清空 is_default。仅限企业所有者操作。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int64 true "企业ID"
// @Success 200 {object} v1.Response
// @Router /enterprise/{id} [delete]
func (h *EnterpriseHandler) Delete(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id == 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "invalid id")
		return
	}
	if err := h.enterpriseService.Delete(ctx, userID, id); err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.Delete error", zap.Error(err))
		switch err {
		case service.ErrForbidden:
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
		case service.ErrEnterpriseNotFound:
			v1.HandleError(ctx, http.StatusNotFound, v1.ErrEnterpriseNotFound, err.Error())
		default:
			v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		}
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// SetDefault godoc
// @Summary 设为默认企业
// @Description 将指定企业设为当前用户的默认企业，同时清除其他企业的默认状态。目标企业必须已认证（status=2）。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.EnterpriseSetDefaultRequest true "params"
// @Success 200 {object} v1.Response
// @Router /enterprise/set_default [post]
func (h *EnterpriseHandler) SetDefault(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	var req v1.EnterpriseSetDefaultRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.enterpriseService.SetDefault(ctx, userID, req.ID); err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.SetDefault error", zap.Error(err))
		switch err {
		case service.ErrForbidden:
			v1.HandleError(ctx, http.StatusForbidden, v1.ErrForbidden, err.Error())
		case service.ErrEnterpriseNotFound:
			v1.HandleError(ctx, http.StatusNotFound, v1.ErrEnterpriseNotFound, err.Error())
		default:
			v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		}
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// My godoc
// @Summary 我的企业列表
// @Description 返回当前用户的所有企业（不含已删除），默认企业排在第一位。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.EnterpriseMyResponseData
// @Router /enterprise/my [get]
func (h *EnterpriseHandler) My(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	list, err := h.enterpriseService.ListByUser(ctx, userID)
	if err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.ListByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	items := make([]v1.EnterpriseMyItem, 0, len(list))
	for _, e := range list {
		items = append(items, v1.EnterpriseMyItem{
			ID:                  e.ID,
			Name:                e.Name,
			SocialCreditCode:    e.SocialCreditCode,
			LegalRepresentative: e.LegalRepresentative,
			Status:              e.Status,
			IsDefault:           e.IsDefault,
		})
	}
	v1.HandleSuccess(ctx, v1.EnterpriseMyResponseData{List: items})
}

// SelectList godoc
// @Summary 发布招聘时选择企业（精简列表）
// @Description 返回当前用户已认证的企业列表，用于发布招聘时弹窗选择，默认企业排在第一位。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} v1.EnterpriseSelectListResponseData
// @Router /enterprise/select_list [get]
func (h *EnterpriseHandler) SelectList(ctx *gin.Context) {
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	list, err := h.enterpriseService.ListVerifiedByUser(ctx, userID)
	if err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.ListVerifiedByUser error", zap.Error(err))
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	items := make([]v1.EnterpriseSelectItem, 0, len(list))
	for _, e := range list {
		items = append(items, v1.EnterpriseSelectItem{
			ID:        e.ID,
			Name:      e.Name,
			IsDefault: e.IsDefault,
		})
	}
	v1.HandleSuccess(ctx, v1.EnterpriseSelectListResponseData{List: items})
}

// Detail godoc
// @Summary 企业详情（公开）
// @Description 返回企业工商信息，无需登录。status≠2（已认证）的企业返回404。不返回营业执照图片URL。
// @Tags 企业管理
// @Accept json
// @Produce json
// @Param id query int true "企业ID"
// @Success 200 {object} v1.EnterpriseDetailResponseData
// @Router /enterprise/detail [get]
func (h *EnterpriseHandler) Detail(ctx *gin.Context) {
	var req v1.EnterpriseDetailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	e, err := h.enterpriseService.GetByID(ctx, req.ID)
	if err != nil {
		h.logger.WithContext(ctx).Error("enterpriseService.GetByID error", zap.Error(err))
		if err == service.ErrEnterpriseNotFound {
			v1.HandleError(ctx, http.StatusNotFound, v1.ErrEnterpriseNotFound, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	if e.Status != model.EnterpriseStatusVerified {
		v1.HandleError(ctx, http.StatusNotFound, v1.ErrEnterpriseNotFound, "enterprise not found")
		return
	}
	v1.HandleSuccess(ctx, v1.EnterpriseDetailResponseData{
		ID:                  e.ID,
		Name:                e.Name,
		LegalRepresentative: e.LegalRepresentative,
		RegisteredCapital:   e.RegisteredCapital,
		EstablishedDate:     e.EstablishedDate,
		BusinessPeriod:      e.BusinessPeriod,
		SocialCreditCode:    e.SocialCreditCode,
		Address:             e.Address,
		BusinessScope:       e.BusinessScope,
	})
}
