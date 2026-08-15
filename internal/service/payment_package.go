package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

var paymentPackageSKURegexp = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,63}$`)

type PaymentPackageCreateInput struct {
	ProductID          int64
	SKUCode            string
	VirtualProductID   string
	Name               string
	Subtitle           string
	Badge              string
	PriceCents         int64
	OriginalPriceCents int64
	Sort               int
	BenefitConfig      model.PaymentBenefitConfig
	SaleRule           model.PaymentSaleRule
	PromotionConfig    model.PaymentPromotionConfig
	Operator           string
}

type PaymentPackageUpdateInput struct {
	ID                 int64
	Version            int
	Name               string
	Subtitle           string
	Badge              string
	VirtualProductID   string
	PriceCents         int64
	OriginalPriceCents int64
	Sort               int
	BenefitConfig      model.PaymentBenefitConfig
	SaleRule           model.PaymentSaleRule
	PromotionConfig    model.PaymentPromotionConfig
	ChangeReason       string
	Operator           string
}

type PaymentPackageAggregate struct {
	Product            *model.PaymentProduct
	Package            *model.PaymentPackage
	BenefitConfig      model.PaymentBenefitConfig
	SaleRule           model.PaymentSaleRule
	PromotionConfig    model.PaymentPromotionConfig
	PurchasePriceCents int64
	Promotion          *model.PaymentPromotionSnapshot
	Purchasable        bool
	UnavailableReason  string
}

type PaymentProductAggregate struct {
	Product      *model.PaymentProduct
	PackageCount int64
}

type PaymentPackageService interface {
	ListProducts(ctx context.Context) ([]PaymentProductAggregate, error)
	UpdateProductPurchaseNotice(ctx context.Context, productID int64, purchaseNotice string) error
	AdminList(ctx context.Context, query repository.PaymentPackageListQuery) ([]PaymentPackageAggregate, int64, error)
	GetByID(ctx context.Context, id int64) (*PaymentPackageAggregate, error)
	Create(ctx context.Context, input PaymentPackageCreateInput) (int64, error)
	Update(ctx context.Context, input PaymentPackageUpdateInput) error
	Delete(ctx context.Context, id int64, version int, reason, operator string) error
	Publish(ctx context.Context, id int64, version int, operator string) error
	Unpublish(ctx context.Context, id int64, version int, reason, operator string) error
	History(ctx context.Context, packageID int64, pageNum, pageSize int) ([]*model.PaymentPackageChangeLog, int64, error)
	ListAvailable(ctx context.Context, userID int64, productCode string, bizType int) (*model.PaymentProduct, []PaymentPackageAggregate, error)
	ResolveForPurchase(ctx context.Context, userID int64, skuCode, productCode string, bizType int) (*PaymentPackageAggregate, error)
}

func NewPaymentPackageService(
	service *Service,
	repository repository.PaymentPackageRepository,
	productRepository repository.PaymentProductRepository,
	userRepository repository.UserRepository,
	orderRepository repository.OrderRepository,
) PaymentPackageService {
	return &paymentPackageService{
		Service:           service,
		repository:        repository,
		productRepository: productRepository,
		userRepository:    userRepository,
		orderRepository:   orderRepository,
	}
}

type paymentPackageService struct {
	*Service
	repository        repository.PaymentPackageRepository
	productRepository repository.PaymentProductRepository
	userRepository    repository.UserRepository
	orderRepository   repository.OrderRepository
}

func (s *paymentPackageService) ListProducts(ctx context.Context) ([]PaymentProductAggregate, error) {
	products, err := s.productRepository.ListWithPackageCount(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]PaymentProductAggregate, 0, len(products))
	for _, product := range products {
		result = append(result, PaymentProductAggregate{
			Product:      product.Product,
			PackageCount: product.PackageCount,
		})
	}
	return result, nil
}

func (s *paymentPackageService) UpdateProductPurchaseNotice(
	ctx context.Context,
	productID int64,
	purchaseNotice string,
) error {
	product, err := s.productRepository.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return ErrPaymentProductNotFound
	}
	return s.productRepository.UpdatePurchaseNotice(ctx, productID, strings.TrimSpace(purchaseNotice))
}

func (s *paymentPackageService) AdminList(
	ctx context.Context,
	query repository.PaymentPackageListQuery,
) ([]PaymentPackageAggregate, int64, error) {
	packages, total, err := s.repository.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	return s.loadAggregates(ctx, packages, total)
}

func (s *paymentPackageService) GetByID(ctx context.Context, id int64) (*PaymentPackageAggregate, error) {
	pkg, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, ErrPaymentPackageNotFound
	}
	product, err := s.getProductForPackage(ctx, pkg)
	if err != nil {
		return nil, err
	}
	return &PaymentPackageAggregate{
		Product:            product,
		Package:            pkg,
		BenefitConfig:      decodeBenefitConfig(pkg.BenefitConfigJSON),
		SaleRule:           decodeSaleRule(pkg.SaleRuleJSON),
		PromotionConfig:    decodePromotionConfig(pkg.PromotionConfigJSON),
		PurchasePriceCents: pkg.PriceCents,
		Purchasable:        true,
	}, nil
}

func (s *paymentPackageService) Create(ctx context.Context, input PaymentPackageCreateInput) (int64, error) {
	input.SKUCode = strings.TrimSpace(input.SKUCode)
	if !paymentPackageSKURegexp.MatchString(input.SKUCode) {
		return 0, ErrPaymentPackageInvalid
	}
	if input.ProductID <= 0 {
		return 0, ErrPaymentProductNotFound
	}
	productID := input.ProductID
	var pkg *model.PaymentPackage
	err := s.tm.Transaction(ctx, func(ctx context.Context) error {
		product, err := s.productRepository.GetByIDForUpdate(ctx, productID)
		if err != nil {
			return err
		}
		if product == nil {
			return ErrPaymentProductNotFound
		}
		count, err := s.repository.CountByProductID(ctx, product.ID)
		if err != nil {
			return err
		}
		if err := validatePackageCardinality(product.SelectionMode, count); err != nil {
			return err
		}
		exists, err := s.repository.ExistsBySKU(ctx, input.SKUCode)
		if err != nil {
			return err
		}
		if exists {
			return ErrPaymentPackageSKUExists
		}
		if err := validatePaymentPackage(
			product.ProductCode,
			input.Name,
			input.PriceCents,
			input.OriginalPriceCents,
			input.BenefitConfig,
		); err != nil {
			return err
		}
		if err := validateSaleRule(input.SaleRule); err != nil {
			return err
		}
		if err := validatePromotionConfig(input.PromotionConfig, input.PriceCents); err != nil {
			return err
		}
		benefitConfigJSON, _ := json.Marshal(input.BenefitConfig)
		saleRuleJSON, _ := json.Marshal(normalizeSaleRule(input.SaleRule))
		promotionConfigJSON, _ := json.Marshal(normalizePromotionConfig(input.PromotionConfig))
		now := time.Now()
		pkg = &model.PaymentPackage{
			ProductID:           product.ID,
			SKUCode:             input.SKUCode,
			VirtualProductID:    strings.TrimSpace(input.VirtualProductID),
			Name:                strings.TrimSpace(input.Name),
			Subtitle:            strings.TrimSpace(input.Subtitle),
			Badge:               strings.TrimSpace(input.Badge),
			PriceCents:          input.PriceCents,
			OriginalPriceCents:  input.OriginalPriceCents,
			BenefitConfigJSON:   string(benefitConfigJSON),
			SaleRuleJSON:        string(saleRuleJSON),
			PromotionConfigJSON: string(promotionConfigJSON),
			Status:              model.PaymentPackageStatusDraft,
			Sort:                input.Sort,
			Version:             1,
			CreatedBy:           input.Operator,
			UpdatedBy:           input.Operator,
			CreateAt:            now,
			UpdateAt:            now,
		}
		if err := s.repository.Create(ctx, pkg); err != nil {
			return err
		}
		after := marshalPaymentPackageSnapshot(pkg)
		return s.repository.CreateChangeLog(ctx, &model.PaymentPackageChangeLog{
			SKUID:          pkg.ID,
			SKUVersion:     pkg.Version,
			Action:         model.PaymentPackageActionCreate,
			BeforeSnapshot: "null",
			AfterSnapshot:  after,
			Operator:       input.Operator,
			CreateAt:       now,
		})
	})
	if err != nil {
		return 0, err
	}
	return pkg.ID, nil
}

func (s *paymentPackageService) Update(ctx context.Context, input PaymentPackageUpdateInput) error {
	current, err := s.GetByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if current.Package.Status == model.PaymentPackageStatusPublished {
		return ErrPaymentPackagePublished
	}
	if current.Package.Version != input.Version {
		return ErrPaymentPackageConflict
	}
	if err := validatePaymentPackage(
		current.Product.ProductCode,
		input.Name,
		input.PriceCents,
		input.OriginalPriceCents,
		input.BenefitConfig,
	); err != nil {
		return err
	}
	if err := validateSaleRule(input.SaleRule); err != nil {
		return err
	}
	if err := validatePromotionConfig(input.PromotionConfig, input.PriceCents); err != nil {
		return err
	}
	benefitConfigJSON, _ := json.Marshal(input.BenefitConfig)
	saleRuleJSON, _ := json.Marshal(normalizeSaleRule(input.SaleRule))
	promotionConfigJSON, _ := json.Marshal(normalizePromotionConfig(input.PromotionConfig))
	now := time.Now()
	updated := *current.Package
	updated.Name = strings.TrimSpace(input.Name)
	updated.Subtitle = strings.TrimSpace(input.Subtitle)
	updated.Badge = strings.TrimSpace(input.Badge)
	updated.VirtualProductID = strings.TrimSpace(input.VirtualProductID)
	updated.PriceCents = input.PriceCents
	updated.OriginalPriceCents = input.OriginalPriceCents
	updated.BenefitConfigJSON = string(benefitConfigJSON)
	updated.SaleRuleJSON = string(saleRuleJSON)
	updated.PromotionConfigJSON = string(promotionConfigJSON)
	updated.Sort = input.Sort
	updated.Version++
	updated.UpdatedBy = input.Operator
	updated.UpdateAt = now
	beforeSnapshot := marshalPaymentPackageSnapshot(current.Package)
	afterSnapshot := marshalPaymentPackageSnapshot(&updated)
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		if _, err := s.productRepository.GetByIDForUpdate(ctx, current.Product.ID); err != nil {
			return err
		}
		ok, err := s.repository.Update(ctx, &updated, input.Version)
		if err != nil {
			return err
		}
		if !ok {
			return ErrPaymentPackageConflict
		}
		return s.repository.CreateChangeLog(ctx, &model.PaymentPackageChangeLog{
			SKUID:          input.ID,
			SKUVersion:     updated.Version,
			Action:         model.PaymentPackageActionUpdate,
			BeforeSnapshot: beforeSnapshot,
			AfterSnapshot:  afterSnapshot,
			ChangeReason:   strings.TrimSpace(input.ChangeReason),
			Operator:       input.Operator,
			CreateAt:       now,
		})
	})
}

func (s *paymentPackageService) Delete(ctx context.Context, id int64, version int, reason, operator string) error {
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current.Package.Status == model.PaymentPackageStatusPublished {
		return ErrPaymentPackagePublished
	}
	if current.Product.SelectionMode == model.PaymentSelectionModeSingle {
		return ErrPaymentPackageSingleDelete
	}
	if current.Package.Version != version {
		return ErrPaymentPackageConflict
	}
	now := time.Now()
	afterPackage := *current.Package
	afterPackage.DeletedAt = &now
	afterPackage.Version++
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		ok, err := s.repository.SoftDelete(ctx, id, version, operator)
		if err != nil {
			return err
		}
		if !ok {
			return ErrPaymentPackageConflict
		}
		return s.repository.CreateChangeLog(ctx, &model.PaymentPackageChangeLog{
			SKUID:          id,
			SKUVersion:     afterPackage.Version,
			Action:         model.PaymentPackageActionDelete,
			BeforeSnapshot: marshalPaymentPackageSnapshot(current.Package),
			AfterSnapshot:  marshalPaymentPackageSnapshot(&afterPackage),
			ChangeReason:   strings.TrimSpace(reason),
			Operator:       operator,
			CreateAt:       now,
		})
	})
}

func (s *paymentPackageService) Publish(ctx context.Context, id int64, version int, operator string) error {
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current.Package.Version != version {
		return ErrPaymentPackageConflict
	}
	if current.Package.Status != model.PaymentPackageStatusDraft &&
		current.Package.Status != model.PaymentPackageStatusUnpublished {
		return ErrPaymentPackageInvalid
	}
	if err := validatePaymentPackage(
		current.Product.ProductCode,
		current.Package.Name,
		current.Package.PriceCents,
		current.Package.OriginalPriceCents,
		current.BenefitConfig,
	); err != nil {
		return err
	}
	if err := validateSaleRule(current.SaleRule); err != nil {
		return err
	}
	if err := validatePromotionConfig(current.PromotionConfig, current.Package.PriceCents); err != nil {
		return err
	}
	virtualProductID := strings.TrimSpace(current.Package.VirtualProductID)
	if virtualProductID == "" {
		return ErrVirtualPaymentUnavailable
	}
	promotionProductID := strings.TrimSpace(current.PromotionConfig.VirtualProductID)
	if current.PromotionConfig.FirstPurchasePriceCents > 0 && promotionProductID == "" {
		return ErrVirtualPaymentUnavailable
	}
	if promotionProductID != "" && promotionProductID == virtualProductID {
		return ErrPaymentPackageInvalid
	}
	exists, err := s.repository.ExistsByVirtualProductID(ctx, virtualProductID, current.Package.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrPaymentPackageInvalid
	}
	if promotionProductID != "" {
		exists, err = s.repository.ExistsByVirtualProductID(ctx, promotionProductID, current.Package.ID)
		if err != nil {
			return err
		}
		if exists {
			return ErrPaymentPackageInvalid
		}
	}
	return s.changeStatus(ctx, current, model.PaymentPackageStatusPublished, model.PaymentPackageActionPublish, "", operator)
}

func (s *paymentPackageService) Unpublish(ctx context.Context, id int64, version int, reason, operator string) error {
	current, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if current.Package.Version != version {
		return ErrPaymentPackageConflict
	}
	if current.Package.Status != model.PaymentPackageStatusPublished {
		return ErrPaymentPackageInvalid
	}
	return s.changeStatus(ctx, current, model.PaymentPackageStatusUnpublished, model.PaymentPackageActionUnpublish, reason, operator)
}

func (s *paymentPackageService) History(
	ctx context.Context,
	packageID int64,
	pageNum, pageSize int,
) ([]*model.PaymentPackageChangeLog, int64, error) {
	return s.repository.ListChangeLogs(ctx, packageID, pageNum, pageSize)
}

func (s *paymentPackageService) ListAvailable(
	ctx context.Context,
	userID int64,
	productCode string,
	bizType int,
) (*model.PaymentProduct, []PaymentPackageAggregate, error) {
	if bizType < 0 || bizType > 3 {
		return nil, nil, ErrPaymentPackageInvalid
	}
	product, err := s.getProductByCode(ctx, productCode)
	if err != nil {
		return nil, nil, err
	}
	packages, err := s.repository.ListAvailable(ctx, product.ID)
	if err != nil {
		return nil, nil, err
	}
	if !supportsBizType(product.ProductCode, bizType) {
		return nil, nil, ErrPaymentPackageInvalid
	}
	result, _, err := s.loadAggregates(ctx, packages, int64(len(packages)))
	if err != nil {
		return nil, nil, err
	}
	if product.SelectionMode == model.PaymentSelectionModeSingle && len(result) > 1 {
		return nil, nil, ErrPaymentPackageCardinality
	}
	available := make([]PaymentPackageAggregate, 0, len(result))
	for _, item := range result {
		// 小程序仅展示已绑定微信道具的 SKU，避免用户选中后无法发起虚拟支付。
		if strings.TrimSpace(item.Package.VirtualProductID) == "" {
			continue
		}
		purchasable, reason, err := s.evaluatePurchaseRule(ctx, userID, item)
		if err != nil {
			return nil, nil, err
		}
		item.Purchasable = purchasable
		item.UnavailableReason = reason
		if !purchasable {
			continue
		}
		if err := s.applyPurchasePromotion(ctx, userID, &item, false); err != nil {
			return nil, nil, err
		}
		available = append(available, item)
	}
	return product, available, nil
}

func (s *paymentPackageService) ResolveForPurchase(
	ctx context.Context,
	userID int64,
	skuCode, productCode string,
	bizType int,
) (*PaymentPackageAggregate, error) {
	skuCode = strings.TrimSpace(skuCode)
	productCode = strings.TrimSpace(productCode)
	if userID <= 0 || !paymentPackageSKURegexp.MatchString(skuCode) || !validPaymentProductCode(productCode) {
		return nil, ErrPaymentPackageInvalid
	}
	pkg, err := s.repository.GetBySKU(ctx, skuCode)
	if err != nil {
		return nil, err
	}
	if pkg == nil {
		return nil, ErrPaymentPackageNotFound
	}
	if pkg.Status != model.PaymentPackageStatusPublished {
		return nil, ErrPaymentPackageUnavailable
	}
	aggregate, err := s.GetByID(ctx, pkg.ID)
	if err != nil {
		return nil, err
	}
	if aggregate.Product.ProductCode != productCode {
		return nil, ErrPaymentPackageInvalid
	}
	if !supportsBizType(productCode, bizType) {
		return nil, ErrPaymentPackageInvalid
	}
	// 所有收费业务已统一使用虚拟支付；未绑定微信道具的 SKU 不允许创建待支付订单。
	if strings.TrimSpace(aggregate.Package.VirtualProductID) == "" {
		return nil, ErrVirtualPaymentUnavailable
	}
	purchasable, reason, err := s.evaluatePurchaseRule(ctx, userID, *aggregate)
	if err != nil {
		return nil, err
	}
	if !purchasable {
		if strings.Contains(reason, "限购") {
			return nil, ErrPaymentPackageLimitReached
		}
		return nil, ErrPaymentPackageUnavailable
	}
	aggregate.Purchasable = true
	if err := s.applyPurchasePromotion(ctx, userID, aggregate, true); err != nil {
		return nil, err
	}
	return aggregate, nil
}

func (s *paymentPackageService) changeStatus(
	ctx context.Context,
	current *PaymentPackageAggregate,
	to model.PaymentPackageStatus,
	action model.PaymentPackageAction,
	reason, operator string,
) error {
	now := time.Now()
	after := *current.Package
	after.Status = to
	after.Version++
	after.UpdatedBy = operator
	after.UpdateAt = now
	return s.tm.Transaction(ctx, func(ctx context.Context) error {
		ok, err := s.repository.UpdateStatus(
			ctx,
			current.Package.ID,
			current.Package.Version,
			current.Package.Status,
			to,
			operator,
		)
		if err != nil {
			return err
		}
		if !ok {
			return ErrPaymentPackageConflict
		}
		return s.repository.CreateChangeLog(ctx, &model.PaymentPackageChangeLog{
			SKUID:          current.Package.ID,
			SKUVersion:     after.Version,
			Action:         action,
			BeforeSnapshot: marshalPaymentPackageSnapshot(current.Package),
			AfterSnapshot:  marshalPaymentPackageSnapshot(&after),
			ChangeReason:   strings.TrimSpace(reason),
			Operator:       operator,
			CreateAt:       now,
		})
	})
}

func (s *paymentPackageService) loadAggregates(
	ctx context.Context,
	packages []*model.PaymentPackage,
	total int64,
) ([]PaymentPackageAggregate, int64, error) {
	result := make([]PaymentPackageAggregate, 0, len(packages))
	products, err := s.productRepository.List(ctx)
	if err != nil {
		return nil, 0, err
	}
	productsByID := make(map[int64]*model.PaymentProduct, len(products))
	for _, product := range products {
		productsByID[product.ID] = product
	}
	for _, pkg := range packages {
		product := productsByID[pkg.ProductID]
		if product == nil {
			return nil, 0, ErrPaymentProductNotFound
		}
		result = append(result, PaymentPackageAggregate{
			Product:            product,
			Package:            pkg,
			BenefitConfig:      decodeBenefitConfig(pkg.BenefitConfigJSON),
			SaleRule:           decodeSaleRule(pkg.SaleRuleJSON),
			PromotionConfig:    decodePromotionConfig(pkg.PromotionConfigJSON),
			PurchasePriceCents: pkg.PriceCents,
			Purchasable:        true,
		})
	}
	return result, total, nil
}

func (s *paymentPackageService) getProductByCode(
	ctx context.Context,
	productCode string,
) (*model.PaymentProduct, error) {
	var (
		product *model.PaymentProduct
		err     error
	)
	if strings.TrimSpace(productCode) != "" {
		product, err = s.productRepository.GetByCode(ctx, strings.TrimSpace(productCode))
	} else {
		return nil, ErrPaymentPackageInvalid
	}
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrPaymentProductNotFound
	}
	return product, nil
}

func (s *paymentPackageService) getProductForPackage(
	ctx context.Context,
	pkg *model.PaymentPackage,
) (*model.PaymentProduct, error) {
	var (
		product *model.PaymentProduct
		err     error
	)
	product, err = s.productRepository.GetByID(ctx, pkg.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, ErrPaymentProductNotFound
	}
	return product, nil
}

func validatePackageCardinality(mode model.PaymentSelectionMode, packageCount int64) error {
	switch mode {
	case model.PaymentSelectionModeSingle:
		if packageCount > 0 {
			return ErrPaymentPackageCardinality
		}
	case model.PaymentSelectionModeMultiple:
	default:
		return ErrPaymentPackageInvalid
	}
	return nil
}

func validatePaymentPackage(
	productCode string,
	name string,
	priceCents, originalPriceCents int64,
	benefits model.PaymentBenefitConfig,
) error {
	if !validPaymentProductCode(productCode) || strings.TrimSpace(name) == "" || priceCents <= 0 {
		return ErrPaymentPackageInvalid
	}
	if originalPriceCents < 0 || (originalPriceCents > 0 && originalPriceCents < priceCents) {
		return ErrPaymentPackageInvalid
	}
	if benefits.ContactVouchers < 0 || benefits.TopHours < 0 || benefits.RefreshTimes < 0 || benefits.RentPublishTimes < 0 || benefits.GiftContactVouchers < 0 {
		return ErrPaymentPackageInvalid
	}
	switch productCode {
	case model.PaymentProductCodeContactVoucher:
		if benefits.ContactVouchers <= 0 || benefits.GiftContactVouchers > 0 || benefits.TopHours > 0 || benefits.RefreshTimes > 0 || benefits.RentPublishTimes > 0 {
			return ErrPaymentPackageInvalid
		}
	case model.PaymentProductCodeJobTop:
		if benefits.TopHours <= 0 || benefits.ContactVouchers > 0 || benefits.RefreshTimes > 0 || benefits.RentPublishTimes > 0 {
			return ErrPaymentPackageInvalid
		}
	case model.PaymentProductCodePaidRefresh:
		if benefits.RefreshTimes != 1 || benefits.ContactVouchers > 0 || benefits.TopHours > 0 || benefits.RentPublishTimes > 0 || benefits.GiftContactVouchers > 0 {
			return ErrPaymentPackageInvalid
		}
	case model.PaymentProductCodeRentPublish:
		if benefits.RentPublishTimes != 1 || benefits.ContactVouchers > 0 || benefits.TopHours > 0 || benefits.RefreshTimes > 0 || benefits.GiftContactVouchers > 0 {
			return ErrPaymentPackageInvalid
		}
	default:
		return ErrPaymentPackageInvalid
	}
	return nil
}

func validateSaleRule(input model.PaymentSaleRule) error {
	if input.MaxPurchasePerUser < 0 {
		return ErrPaymentPackageInvalid
	}
	return nil
}

func validatePromotionConfig(input model.PaymentPromotionConfig, regularPriceCents int64) error {
	if input.FirstPurchasePriceCents == 0 && input.FirstPurchaseScope == "" && strings.TrimSpace(input.Subtitle) == "" && strings.TrimSpace(input.Badge) == "" && strings.TrimSpace(input.VirtualProductID) == "" {
		return nil
	}
	if input.FirstPurchasePriceCents <= 0 || input.FirstPurchasePriceCents >= regularPriceCents {
		return ErrPaymentPackageInvalid
	}
	if input.FirstPurchaseScope != model.PaymentFirstPurchaseScopePlatform && input.FirstPurchaseScope != model.PaymentFirstPurchaseScopeProduct {
		return ErrPaymentPackageInvalid
	}
	if strings.TrimSpace(input.VirtualProductID) == "" {
		return ErrPaymentPackageInvalid
	}
	return nil
}

func validPaymentProductCode(productCode string) bool {
	switch productCode {
	case model.PaymentProductCodeJobTop, model.PaymentProductCodeContactVoucher, model.PaymentProductCodePaidRefresh, model.PaymentProductCodeRentPublish:
		return true
	default:
		return false
	}
}

func supportsBizType(productCode string, bizType int) bool {
	switch productCode {
	case model.PaymentProductCodeContactVoucher:
		return bizType == 0
	case model.PaymentProductCodeJobTop:
		return bizType == 1 || bizType == 2
	case model.PaymentProductCodePaidRefresh:
		return bizType >= 1 && bizType <= 3
	case model.PaymentProductCodeRentPublish:
		return bizType == 3
	default:
		return false
	}
}

func decodeBenefitConfig(raw string) model.PaymentBenefitConfig {
	var config model.PaymentBenefitConfig
	_ = json.Unmarshal([]byte(raw), &config)
	return config
}

func decodeSaleRule(raw string) model.PaymentSaleRule {
	var rule model.PaymentSaleRule
	_ = json.Unmarshal([]byte(raw), &rule)
	return normalizeSaleRule(rule)
}

func normalizeSaleRule(rule model.PaymentSaleRule) model.PaymentSaleRule {
	return rule
}

func decodePromotionConfig(raw string) model.PaymentPromotionConfig {
	var config model.PaymentPromotionConfig
	_ = json.Unmarshal([]byte(raw), &config)
	return normalizePromotionConfig(config)
}

func normalizePromotionConfig(config model.PaymentPromotionConfig) model.PaymentPromotionConfig {
	config.Subtitle = strings.TrimSpace(config.Subtitle)
	config.Badge = strings.TrimSpace(config.Badge)
	config.VirtualProductID = strings.TrimSpace(config.VirtualProductID)
	if config.FirstPurchasePriceCents == 0 {
		config.FirstPurchaseScope = ""
		config.Subtitle = ""
		config.Badge = ""
		config.VirtualProductID = ""
	}
	return config
}

// applyPurchasePromotion resolves the current user's price for one SKU. It is
// deliberately run again while creating an order, so an old list response can
// never grant a discount after the user's first purchase has been created.
func (s *paymentPackageService) applyPurchasePromotion(
	ctx context.Context,
	userID int64,
	item *PaymentPackageAggregate,
	lockUser bool,
) error {
	item.PurchasePriceCents = item.Package.PriceCents
	item.Promotion = nil
	config := item.PromotionConfig
	if config.FirstPurchasePriceCents == 0 {
		return nil
	}
	if userID <= 0 || s.userRepository == nil || s.orderRepository == nil {
		return ErrPaymentPackageInvalid
	}
	if lockUser {
		user, err := s.userRepository.GetByIDForUpdate(ctx, userID)
		if err != nil {
			return err
		}
		if user == nil {
			return ErrPaymentPackageInvalid
		}
	}

	var (
		count int64
		err   error
	)
	switch config.FirstPurchaseScope {
	case model.PaymentFirstPurchaseScopePlatform:
		count, err = s.orderRepository.CountActiveOrders(ctx, userID)
	case model.PaymentFirstPurchaseScopeProduct:
		count, err = s.orderRepository.CountActivePurchases(ctx, userID, item.Product.ID, 0, nil)
	default:
		return ErrPaymentPackageInvalid
	}
	if err != nil {
		return err
	}
	if count != 0 {
		return nil
	}
	item.PurchasePriceCents = config.FirstPurchasePriceCents
	item.Promotion = &model.PaymentPromotionSnapshot{
		Type:                "first_purchase",
		Scope:               config.FirstPurchaseScope,
		RegularPriceCents:   item.Package.PriceCents,
		PromotionPriceCents: config.FirstPurchasePriceCents,
		Subtitle:            config.Subtitle,
		Badge:               config.Badge,
		VirtualProductID:    config.VirtualProductID,
	}
	return nil
}

func (s *paymentPackageService) evaluatePurchaseRule(
	ctx context.Context,
	userID int64,
	item PaymentPackageAggregate,
) (bool, string, error) {
	rule := item.SaleRule
	if rule.MaxPurchasePerUser == 0 {
		return true, "", nil
	}
	if userID <= 0 {
		return false, "请登录后购买", nil
	}
	if s.orderRepository == nil {
		return false, "", ErrPaymentPackageInvalid
	}
	total, err := s.orderRepository.CountActivePurchases(ctx, userID, 0, item.Package.ID, nil)
	if err != nil {
		return false, "", err
	}
	if total >= int64(rule.MaxPurchasePerUser) {
		return false, "已达到每人累计限购次数", nil
	}
	return true, "", nil
}

func marshalPaymentPackageSnapshot(pkg *model.PaymentPackage) string {
	data, err := json.Marshal(pkg)
	if err != nil {
		return "null"
	}
	return string(data)
}
