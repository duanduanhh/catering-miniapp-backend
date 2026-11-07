package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type PaymentPackageHandler struct {
	*Handler
	service service.PaymentPackageService
}

func NewPaymentPackageHandler(handler *Handler, service service.PaymentPackageService) *PaymentPackageHandler {
	return &PaymentPackageHandler{Handler: handler, service: service}
}

// AdminProducts godoc
// @Summary 管理后台付费产品列表
// @Description 返回收费业务、单/多规格模式、购买须知及当前未删除 SKU 数量。selection_mode：1=单规格，2=多规格。
// @Tags 管理后台-付费套餐
// @Produce json
// @Success 200 {object} v1.PaymentProductListResponseData
// @Router /admin/payment-products/list [post]
func (h *PaymentPackageHandler) AdminProducts(ctx *gin.Context) {
	result, err := h.service.ListProducts(ctx)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response := v1.PaymentProductListResponseData{
		List: make([]v1.PaymentProductItem, 0, len(result)),
	}
	for _, item := range result {
		response.List = append(response.List, v1.PaymentProductItem{
			ID:             item.Product.ID,
			ProductCode:    item.Product.ProductCode,
			Name:           item.Product.Name,
			SelectionMode:  int(item.Product.SelectionMode),
			PurchaseNotice: item.Product.PurchaseNotice,
			PackageCount:   item.PackageCount,
			CreateAt:       formatTime(item.Product.CreateAt),
			UpdateAt:       formatTime(item.Product.UpdateAt),
		})
	}
	v1.HandleSuccess(ctx, response)
}

// AdminUpdateProduct godoc
// @Summary 编辑付费产品购买须知
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentProductUpdateRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/payment-products/update [post]
func (h *PaymentPackageHandler) AdminUpdateProduct(ctx *gin.Context) {
	var req v1.AdminPaymentProductUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	if err := h.service.UpdateProductPurchaseNotice(ctx, req.ID, req.PurchaseNotice); err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminList godoc
// @Summary 管理后台套餐列表
// @Description 查询全部未删除 SKU，支持按收费业务、状态和关键词筛选。响应中的 price_cents、original_price_cents 单位为分；sale_rule、promotion_config 用于管理端编辑回显。
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageListRequest true "params"
// @Success 200 {object} v1.AdminPaymentPackageListResponseData
// @Router /admin/payment-packages/list [post]
func (h *PaymentPackageHandler) AdminList(ctx *gin.Context) {
	var req v1.AdminPaymentPackageListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	result, total, err := h.service.AdminList(ctx, repository.PaymentPackageListQuery{
		ProductID: req.ProductID,
		Status:    req.Status,
		Keyword:   req.Keyword,
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
	})
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response := v1.AdminPaymentPackageListResponseData{
		List:  make([]v1.PaymentPackageItem, 0, len(result)),
		Total: total,
	}
	for _, item := range result {
		response.List = append(response.List, toPaymentPackageItem(item, true))
	}
	v1.HandleSuccess(ctx, response)
}

// AdminDetail godoc
// @Summary 管理后台套餐详情
// @Description 返回 SKU 的完整管理配置，包括权益、累计限购和首购特惠配置；promotion_config.first_purchase_price_cents 为同一 SKU 的首购成交价（分），first_purchase_scope：platform=平台首购，product=当前产品首购。
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageDetailRequest true "params"
// @Success 200 {object} v1.PaymentPackageItem
// @Router /admin/payment-packages/detail [post]
func (h *PaymentPackageHandler) AdminDetail(ctx *gin.Context) {
	var req v1.AdminPaymentPackageDetailRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	result, err := h.service.GetByID(ctx, req.ID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, toPaymentPackageItem(*result, true))
}

// AdminCreate godoc
// @Summary 新增付费套餐
// @Description 新增 SKU 后状态为草稿。product_id 和 sku_code 创建后不可修改；金额字段单位为分。上架前必须填写微信虚拟支付后台已发布的 virtual_product_id，且一个道具 ID 只能绑定一个 SKU。promotion_config.first_purchase_price_cents 可为同一 SKU 配置首购特惠价，必须低于 price_cents；first_purchase_scope：platform=平台首购，product=当前产品首购；首购价与常规价不同，必须配置 promotion_config.virtual_product_id。
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageCreateRequest true "params"
// @Success 200 {object} v1.AdminPaymentPackageCreateResponseData
// @Router /admin/payment-packages/create [post]
func (h *PaymentPackageHandler) AdminCreate(ctx *gin.Context) {
	var req v1.AdminPaymentPackageCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	id, err := h.service.Create(ctx, service.PaymentPackageCreateInput{
		ProductID:          req.ProductID,
		SKUCode:            req.SKUCode,
		VirtualProductID:   req.VirtualProductID,
		Name:               req.Name,
		Subtitle:           req.Subtitle,
		Badge:              req.Badge,
		PriceCents:         req.PriceCents,
		OriginalPriceCents: req.OriginalPriceCents,
		Sort:               req.Sort,
		BenefitConfig:      toPaymentBenefitConfig(req.BenefitConfig),
		SaleRule:           toPaymentSaleRule(req.SaleRule),
		PromotionConfig:    toPaymentPromotionConfig(req.PromotionConfig),
		Operator:           middleware.GetAdminUsernameFromCtx(ctx),
	})
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, v1.AdminPaymentPackageCreateResponseData{ID: id})
}

// AdminUpdate godoc
// @Summary 编辑付费套餐
// @Description 仅草稿或已下架 SKU 可编辑，version 用于防止多人覆盖修改。金额字段单位为分；营销规则会随本次编辑整体更新。
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageUpdateRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/payment-packages/update [post]
func (h *PaymentPackageHandler) AdminUpdate(ctx *gin.Context) {
	var req v1.AdminPaymentPackageUpdateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	err := h.service.Update(ctx, service.PaymentPackageUpdateInput{
		ID:                 req.ID,
		Version:            req.Version,
		Name:               req.Name,
		Subtitle:           req.Subtitle,
		Badge:              req.Badge,
		VirtualProductID:   req.VirtualProductID,
		PriceCents:         req.PriceCents,
		OriginalPriceCents: req.OriginalPriceCents,
		Sort:               req.Sort,
		BenefitConfig:      toPaymentBenefitConfig(req.BenefitConfig),
		SaleRule:           toPaymentSaleRule(req.SaleRule),
		PromotionConfig:    toPaymentPromotionConfig(req.PromotionConfig),
		ChangeReason:       req.ChangeReason,
		Operator:           middleware.GetAdminUsernameFromCtx(ctx),
	})
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminDelete godoc
// @Summary 删除付费套餐
// @Description 仅未上架套餐可删除，执行软删除并保留变更记录。
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageIDRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/payment-packages/delete [post]
func (h *PaymentPackageHandler) AdminDelete(ctx *gin.Context) {
	var req v1.AdminPaymentPackageIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Version <= 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "id and version are required")
		return
	}
	err := h.service.Delete(ctx, req.ID, req.Version, req.Reason, middleware.GetAdminUsernameFromCtx(ctx))
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminPublish godoc
// @Summary 上架付费套餐
// @Description 上架前校验套餐权益、价格以及微信虚拟支付道具 ID；未填写道具 ID 或与其他未删除 SKU 重复时不能上架。
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageIDRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/payment-packages/publish [post]
func (h *PaymentPackageHandler) AdminPublish(ctx *gin.Context) {
	var req v1.AdminPaymentPackageIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Version <= 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "id and version are required")
		return
	}
	err := h.service.Publish(ctx, req.ID, req.Version, middleware.GetAdminUsernameFromCtx(ctx))
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminUnpublish godoc
// @Summary 下架付费套餐
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageIDRequest true "params"
// @Success 200 {object} v1.Response
// @Router /admin/payment-packages/unpublish [post]
func (h *PaymentPackageHandler) AdminUnpublish(ctx *gin.Context) {
	var req v1.AdminPaymentPackageIDRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Version <= 0 {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "id and version are required")
		return
	}
	err := h.service.Unpublish(ctx, req.ID, req.Version, req.Reason, middleware.GetAdminUsernameFromCtx(ctx))
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	v1.HandleSuccess(ctx, nil)
}

// AdminHistory godoc
// @Summary 查询套餐变更记录
// @Tags 管理后台-付费套餐
// @Accept json
// @Produce json
// @Param request body v1.AdminPaymentPackageHistoryRequest true "params"
// @Success 200 {object} v1.AdminPaymentPackageHistoryResponseData
// @Router /admin/payment-packages/history [post]
func (h *PaymentPackageHandler) AdminHistory(ctx *gin.Context) {
	var req v1.AdminPaymentPackageHistoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	list, total, err := h.service.History(ctx, req.SKUID, req.PageNum, req.PageSize)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response := v1.AdminPaymentPackageHistoryResponseData{
		List:  make([]v1.AdminPaymentPackageHistoryItem, 0, len(list)),
		Total: total,
	}
	for _, item := range list {
		response.List = append(response.List, v1.AdminPaymentPackageHistoryItem{
			ID:             item.ID,
			SKUID:          item.SKUID,
			SKUVersion:     item.SKUVersion,
			Action:         int(item.Action),
			BeforeSnapshot: item.BeforeSnapshot,
			AfterSnapshot:  item.AfterSnapshot,
			ChangeReason:   item.ChangeReason,
			Operator:       item.Operator,
			CreateAt:       formatTime(item.CreateAt),
		})
	}
	v1.HandleSuccess(ctx, response)
}

// ListAvailable godoc
// @Summary 查询可购买套餐
// @Description 小程序登录接口。按产品返回已上架、适用于当前 biz_type 且符合当前用户购买资格的 SKU。
// @Description product_code：job_top=岗位置顶，contact_voucher=联系券，paid_refresh=付费刷新，rent_publish=招租发布。
// @Description biz_type：0=不限，1=招聘，2=求职，3=招租。岗位置顶传1或2；联系券传0；付费刷新传当前信息类型；招租发布传3。
// @Description selection_mode：1=单规格，正常情况下 skus 只有一条，前端直接使用 skus[0]；2=多规格，前端展示 skus 供用户选择。
// @Description price_cents 和 original_price_cents 的单位均为分，例如199表示1.99元。命中首购特惠时，同一 sku_code 返回特惠成交价 price_cents，original_price_cents 返回常规售价。skus 为空表示当前没有可购买 SKU。
// @Description benefit_config 包含联系券、刷新次数、置顶小时数、招租发布次数和赠送联系券；仅返回当前产品适用的字段。营销规则由服务端完成资格判断，不返回 sale_rule。
// @Description 创建支付订单时只传服务端返回的 sku_code 和业务目标，不能传客户端价格。
// @Tags 付费套餐
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body v1.PaymentPackageListRequest true "产品编码及当前业务类型"
// @Success 200 {object} v1.Response{data=v1.PaymentPackageListResponseData} "查询成功"
// @Failure 400 {object} v1.Response "参数错误，例如 biz_type 超出0～3"
// @Failure 401 {object} v1.Response "未登录或登录态已失效"
// @Failure 404 {object} v1.Response "product_code 对应的产品不存在"
// @Failure 500 {object} v1.Response "服务器内部错误"
// @Router /payment-packages/list [post]
func (h *PaymentPackageHandler) ListAvailable(ctx *gin.Context) {
	var req v1.PaymentPackageListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	userID := GetUserIdFromCtx(ctx)
	if userID == 0 {
		v1.HandleError(ctx, http.StatusUnauthorized, v1.ErrUnauthorized, v1.ErrUnauthorized.Error())
		return
	}
	product, result, err := h.service.ListAvailable(ctx, userID, req.ProductCode, req.BizType)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response := v1.PaymentPackageListResponseData{
		ProductCode:    product.ProductCode,
		ProductName:    product.Name,
		PurchaseNotice: product.PurchaseNotice,
		SelectionMode:  int(product.SelectionMode),
		SKUs:           make([]v1.PaymentSKU, 0, len(result)),
	}
	for _, item := range result {
		response.SKUs = append(response.SKUs, v1.PaymentSKU{
			SKUCode:            item.Package.SKUCode,
			Name:               item.Package.Name,
			Subtitle:           purchaseSubtitle(item),
			Badge:              purchaseBadge(item),
			PromotionType:      purchasePromotionType(item),
			PriceCents:         item.PurchasePriceCents,
			OriginalPriceCents: purchaseOriginalPrice(item),
			BenefitConfig:      toV1PaymentBenefitConfig(item.BenefitConfig),
		})
	}
	v1.HandleSuccess(ctx, response)
}

func (h *PaymentPackageHandler) handleError(ctx *gin.Context, err error) {
	h.logger.WithContext(ctx).Error("payment package error", zap.Error(err))
	switch {
	case errors.Is(err, service.ErrPaymentPackageNotFound):
		v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, err.Error())
	case errors.Is(err, service.ErrPaymentPackageInvalid),
		errors.Is(err, service.ErrPaymentPackagePublished),
		errors.Is(err, service.ErrPaymentPackageSKUExists),
		errors.Is(err, service.ErrPaymentPackageCardinality),
		errors.Is(err, service.ErrPaymentPackageSingleDelete),
		errors.Is(err, service.ErrPaymentPackageUnavailable),
		errors.Is(err, service.ErrPaymentPackageLimitReached):
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
	case errors.Is(err, service.ErrPaymentProductNotFound):
		v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, err.Error())
	case errors.Is(err, service.ErrPaymentPackageConflict):
		v1.HandleError(ctx, http.StatusConflict, v1.ErrBadRequest, err.Error())
	default:
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
	}
}

func toPaymentPackageItem(value service.PaymentPackageAggregate, admin bool) v1.PaymentPackageItem {
	pkg := value.Package
	item := v1.PaymentPackageItem{
		ProductID:          pkg.ProductID,
		ProductCode:        value.Product.ProductCode,
		ProductName:        value.Product.Name,
		SKUCode:            pkg.SKUCode,
		VirtualProductID:   pkg.VirtualProductID,
		SelectionMode:      int(value.Product.SelectionMode),
		Name:               pkg.Name,
		Subtitle:           pkg.Subtitle,
		Badge:              pkg.Badge,
		PriceCents:         pkg.PriceCents,
		OriginalPriceCents: pkg.OriginalPriceCents,
		Sort:               pkg.Sort,
		BenefitConfig:      toV1PaymentBenefitConfig(value.BenefitConfig),
	}
	if admin {
		item.ID = pkg.ID
		item.SaleRule = toV1PaymentSaleRule(value.SaleRule)
		item.PromotionConfig = toV1PaymentPromotionConfig(value.PromotionConfig)
		item.Status = int(pkg.Status)
		item.Version = pkg.Version
		item.CreatedBy = pkg.CreatedBy
		item.UpdatedBy = pkg.UpdatedBy
		item.CreateAt = formatTime(pkg.CreateAt)
		item.UpdateAt = formatTime(pkg.UpdateAt)
	}
	return item
}

func toV1PaymentBenefitConfig(value model.PaymentBenefitConfig) v1.PaymentBenefitConfig {
	return v1.PaymentBenefitConfig{
		ContactVouchers: value.ContactVouchers, TopHours: value.TopHours,
		RefreshTimes: value.RefreshTimes, RentPublishTimes: value.RentPublishTimes,
		GiftContactVouchers: value.GiftContactVouchers,
	}
}

func toV1PaymentSaleRule(value model.PaymentSaleRule) v1.PaymentSaleRule {
	return v1.PaymentSaleRule{
		MaxPurchasePerUser: value.MaxPurchasePerUser,
	}
}

func toV1PaymentPromotionConfig(value model.PaymentPromotionConfig) v1.PaymentPromotionConfig {
	return v1.PaymentPromotionConfig{FirstPurchasePriceCents: value.FirstPurchasePriceCents, FirstPurchaseScope: value.FirstPurchaseScope, Subtitle: value.Subtitle, Badge: value.Badge, VirtualProductID: value.VirtualProductID}
}

func toPaymentBenefitConfig(value v1.PaymentBenefitConfig) model.PaymentBenefitConfig {
	return model.PaymentBenefitConfig{ContactVouchers: value.ContactVouchers, TopHours: value.TopHours, RefreshTimes: value.RefreshTimes, RentPublishTimes: value.RentPublishTimes, GiftContactVouchers: value.GiftContactVouchers}
}

func toPaymentSaleRule(value v1.PaymentSaleRule) model.PaymentSaleRule {
	return model.PaymentSaleRule{
		MaxPurchasePerUser: value.MaxPurchasePerUser,
	}
}

func toPaymentPromotionConfig(value v1.PaymentPromotionConfig) model.PaymentPromotionConfig {
	return model.PaymentPromotionConfig{FirstPurchasePriceCents: value.FirstPurchasePriceCents, FirstPurchaseScope: value.FirstPurchaseScope, Subtitle: value.Subtitle, Badge: value.Badge, VirtualProductID: value.VirtualProductID}
}

func purchaseBadge(item service.PaymentPackageAggregate) string {
	if item.Promotion != nil {
		return item.Promotion.Badge
	}
	return item.Package.Badge
}

func purchaseSubtitle(item service.PaymentPackageAggregate) string {
	if item.Promotion != nil && item.Promotion.Subtitle != "" {
		return item.Promotion.Subtitle
	}
	return item.Package.Subtitle
}

func purchasePromotionType(item service.PaymentPackageAggregate) string {
	if item.Promotion != nil {
		return item.Promotion.Type
	}
	return ""
}

func purchaseOriginalPrice(item service.PaymentPackageAggregate) int64 {
	if item.Promotion != nil {
		return item.Promotion.RegularPriceCents
	}
	return item.Package.OriginalPriceCents
}

func parsePaymentPackageTimes(startRaw, endRaw string) (*time.Time, *time.Time, error) {
	start, err := parsePaymentPackageTime(startRaw)
	if err != nil {
		return nil, nil, err
	}
	end, err := parsePaymentPackageTime(endRaw)
	if err != nil {
		return nil, nil, err
	}
	return start, end, nil
}

func parsePaymentPackageTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05", time.RFC3339} {
		parsed, err := time.ParseInLocation(layout, raw, time.Local)
		if err == nil {
			return &parsed, nil
		}
	}
	return nil, errors.New("invalid datetime format")
}

func decodePaymentPackageBizTypes(raw string) []int {
	var result []int
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return []int{}
	}
	return result
}
