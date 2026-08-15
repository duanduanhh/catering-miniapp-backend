package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/repository"
)

func TestValidatePaymentPackageSupportsFourBusinesses(t *testing.T) {
	tests := []struct {
		name        string
		productCode string
		bizTypes    []int
		benefits    model.PaymentBenefitConfig
	}{
		{
			name:        "top",
			productCode: model.PaymentProductCodeJobTop,
			bizTypes:    []int{1, 2},
			benefits:    model.PaymentBenefitConfig{TopHours: 24, GiftContactVouchers: 2},
		},
		{
			name:        "contact voucher",
			productCode: model.PaymentProductCodeContactVoucher,
			benefits:    model.PaymentBenefitConfig{ContactVouchers: 10},
		},
		{
			name:        "refresh",
			productCode: model.PaymentProductCodePaidRefresh,
			bizTypes:    []int{1, 2, 3},
			benefits:    model.PaymentBenefitConfig{RefreshTimes: 1},
		},
		{
			name:        "rent publish",
			productCode: model.PaymentProductCodeRentPublish,
			bizTypes:    []int{3},
			benefits:    model.PaymentBenefitConfig{RentPublishTimes: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePaymentPackage(
				tt.productCode,
				"测试套餐",
				100,
				200,
				tt.benefits,
			)
			if err != nil {
				t.Fatalf("expected valid package, got %v", err)
			}
		})
	}
}

func TestValidatePaymentPackageRejectsInvalidConfiguration(t *testing.T) {
	now := time.Now()
	beforeNow := now.Add(-time.Hour)
	tests := []struct {
		name        string
		productCode string
		price       int64
		original    int64
		bizTypes    []int
		startAt     *time.Time
		endAt       *time.Time
		benefits    model.PaymentBenefitConfig
	}{
		{
			name:        "zero price",
			productCode: model.PaymentProductCodeContactVoucher,
			price:       0,
			benefits:    model.PaymentBenefitConfig{ContactVouchers: 1},
		},
		{
			name:        "original price below sale price",
			productCode: model.PaymentProductCodeContactVoucher,
			price:       200,
			original:    100,
			benefits:    model.PaymentBenefitConfig{ContactVouchers: 1},
		},
		{
			name:        "end before start",
			productCode: model.PaymentProductCodeContactVoucher,
			price:       100,
			startAt:     &now,
			endAt:       &beforeNow,
			benefits:    model.PaymentBenefitConfig{ContactVouchers: 1},
		},
		{
			name:        "top cannot apply to rent",
			productCode: model.PaymentProductCodeJobTop,
			price:       100,
			bizTypes:    []int{3},
			benefits:    model.PaymentBenefitConfig{TopHours: 24},
		},
		{
			name:        "contact voucher is global",
			productCode: model.PaymentProductCodeContactVoucher,
			price:       100,
			bizTypes:    []int{1},
			benefits:    model.PaymentBenefitConfig{ContactVouchers: 1},
		},
		{
			name:        "contact voucher cannot contain gift voucher",
			productCode: model.PaymentProductCodeContactVoucher,
			price:       100,
			benefits:    model.PaymentBenefitConfig{ContactVouchers: 1, GiftContactVouchers: 1},
		},
		{
			name:        "refresh requires scope",
			productCode: model.PaymentProductCodePaidRefresh,
			price:       100,
			benefits:    model.PaymentBenefitConfig{RefreshTimes: 1},
		},
		{
			name:        "rent publish count is fixed to one",
			productCode: model.PaymentProductCodeRentPublish,
			price:       100,
			bizTypes:    []int{3},
			benefits:    model.PaymentBenefitConfig{RentPublishTimes: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePaymentPackage(
				tt.productCode,
				"测试套餐",
				tt.price,
				tt.original,
				tt.benefits,
			)
			if tt.name == "end before start" || tt.name == "top cannot apply to rent" || tt.name == "contact voucher is global" || tt.name == "refresh requires scope" {
				if err != nil {
					t.Fatalf("expected product-defined scope to be valid, got %v", err)
				}
				return
			}
			if !errors.Is(err, ErrPaymentPackageInvalid) {
				t.Fatalf("expected ErrPaymentPackageInvalid, got %v", err)
			}
		})
	}
}

func TestValidatePackageCardinality(t *testing.T) {
	tests := []struct {
		name         string
		mode         model.PaymentSelectionMode
		packageCount int64
		wantErr      error
	}{
		{
			name:         "single product accepts first package",
			mode:         model.PaymentSelectionModeSingle,
			packageCount: 0,
		},
		{
			name:         "single product rejects second package",
			mode:         model.PaymentSelectionModeSingle,
			packageCount: 1,
			wantErr:      ErrPaymentPackageCardinality,
		},
		{
			name:         "multiple product accepts more packages",
			mode:         model.PaymentSelectionModeMultiple,
			packageCount: 10,
		},
		{
			name:         "unknown mode is rejected",
			mode:         0,
			packageCount: 0,
			wantErr:      ErrPaymentPackageInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePackageCardinality(tt.mode, tt.packageCount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestValidatePromotionConfigForFirstPurchasePrice(t *testing.T) {
	valid := model.PaymentPromotionConfig{
		FirstPurchasePriceCents: 190,
		FirstPurchaseScope:      model.PaymentFirstPurchaseScopeProduct,
		VirtualProductID:         "contact_voucher_2_first",
	}
	if err := validatePromotionConfig(valid, 500); err != nil {
		t.Fatalf("valid first purchase promotion: %v", err)
	}

	for _, rule := range []model.PaymentPromotionConfig{
		{FirstPurchasePriceCents: 500, FirstPurchaseScope: model.PaymentFirstPurchaseScopeProduct},
		{FirstPurchasePriceCents: 190},
		{FirstPurchaseScope: model.PaymentFirstPurchaseScopePlatform},
		{FirstPurchasePriceCents: 190, FirstPurchaseScope: "sku"},
	} {
		if err := validatePromotionConfig(rule, 500); !errors.Is(err, ErrPaymentPackageInvalid) {
			t.Fatalf("invalid first purchase promotion %+v: %v", rule, err)
		}
	}
}

func TestPaymentPackageServiceEnforcesProductSelectionMode(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(
		&model.PaymentProduct{},
		&model.PaymentPackage{},
		&model.PaymentPackageChangeLog{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	now := time.Now()
	singleProduct := &model.PaymentProduct{
		ProductCode:   model.PaymentProductCodePaidRefresh,
		Name:          "付费刷新",
		SelectionMode: model.PaymentSelectionModeSingle,
		CreateAt:      now,
		UpdateAt:      now,
	}
	multipleProduct := &model.PaymentProduct{
		ProductCode:   model.PaymentProductCodeContactVoucher,
		Name:          "联系券",
		SelectionMode: model.PaymentSelectionModeMultiple,
		CreateAt:      now,
		UpdateAt:      now,
	}
	if err := db.Create([]*model.PaymentProduct{singleProduct, multipleProduct}).Error; err != nil {
		t.Fatalf("create products: %v", err)
	}

	baseRepository := repository.NewRepository(nil, db)
	packageRepository := repository.NewPaymentPackageRepository(baseRepository)
	productRepository := repository.NewPaymentProductRepository(baseRepository)
	baseService := &Service{tm: repository.NewTransaction(baseRepository)}
	paymentService := NewPaymentPackageService(baseService, packageRepository, productRepository, nil, nil)
	ctx := context.Background()

	createRefresh := func(sku string) error {
		_, err := paymentService.Create(ctx, PaymentPackageCreateInput{
			ProductID:     singleProduct.ID,
			SKUCode:       sku,
			Name:          "刷新1次",
			PriceCents:    199,
			BenefitConfig: model.PaymentBenefitConfig{RefreshTimes: 1},
			Operator:      "test",
		})
		return err
	}
	if err := createRefresh("paid_refresh_1"); err != nil {
		t.Fatalf("create first single package: %v", err)
	}
	if err := createRefresh("paid_refresh_2"); !errors.Is(err, ErrPaymentPackageCardinality) {
		t.Fatalf("expected cardinality error, got %v", err)
	}

	for _, sku := range []string{"contact_voucher_1", "contact_voucher_5"} {
		if _, err := paymentService.Create(ctx, PaymentPackageCreateInput{
			ProductID:     multipleProduct.ID,
			SKUCode:       sku,
			Name:          sku,
			PriceCents:    399,
			BenefitConfig: model.PaymentBenefitConfig{ContactVouchers: 1},
			Operator:      "test",
		}); err != nil {
			t.Fatalf("create multiple package %s: %v", sku, err)
		}
	}
}

func TestPaymentPackageServiceEnforcesPurchaseLimit(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(
		&model.PaymentProduct{},
		&model.PaymentPackage{},
		&model.PaymentPackageChangeLog{},
		&model.User{},
		&model.Order{},
		&model.OrderItem{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	now := time.Now()
	product := &model.PaymentProduct{
		ProductCode:   model.PaymentProductCodeContactVoucher,
		Name:          "联系券",
		SelectionMode: model.PaymentSelectionModeMultiple,
		CreateAt:      now,
		UpdateAt:      now,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	user := &model.User{WechatOpenID: "openid", CreateAt: now, UpdateAt: now}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseRepository := repository.NewRepository(nil, db)
	packageRepository := repository.NewPaymentPackageRepository(baseRepository)
	productRepository := repository.NewPaymentProductRepository(baseRepository)
	userRepository := repository.NewUserRepository(baseRepository)
	orderRepository := repository.NewOrderRepository(baseRepository)
	baseService := &Service{tm: repository.NewTransaction(baseRepository)}
	paymentService := NewPaymentPackageService(
		baseService,
		packageRepository,
		productRepository,
		userRepository,
		orderRepository,
	)
	packageID, err := paymentService.Create(context.Background(), PaymentPackageCreateInput{
		ProductID:        product.ID,
		SKUCode:          "contact_voucher_2_new",
		Name:             "2张联系券",
		PriceCents:       150,
		VirtualProductID: "contact_voucher_2_new",
		BenefitConfig:    model.PaymentBenefitConfig{ContactVouchers: 2},
		SaleRule:         model.PaymentSaleRule{MaxPurchasePerUser: 1},
		Operator:         "test",
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	aggregate, err := paymentService.GetByID(context.Background(), packageID)
	if err != nil {
		t.Fatalf("get package: %v", err)
	}
	if err := paymentService.Publish(context.Background(), packageID, aggregate.Package.Version, "test"); err != nil {
		t.Fatalf("publish package: %v", err)
	}
	if _, err := paymentService.ResolveForPurchase(
		context.Background(),
		user.ID,
		"contact_voucher_2_new",
		model.PaymentProductCodeContactVoucher,
		0,
	); err != nil {
		t.Fatalf("first purchase should be eligible: %v", err)
	}

	order := &model.Order{
		OrderNo:     "PAID001",
		UserID:      user.ID,
		AmountTotal: model.NewDecimalFromFloat64(1.5),
		AmountPaid:  model.NewDecimalFromFloat64(1.5),
		Currency:    "CNY",
		Status:      model.OrderStatusPaid,
		CreateAt:    now,
		UpdateAt:    now,
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	if err := db.Create(&model.OrderItem{
		OrderID:     order.ID,
		ProductType: model.ProductTypeContactVoucher,
		ProductID:   product.ID,
		SKUID:       packageID,
		SKUCode:     "contact_voucher_2_new",
		CreateAt:    now,
		UpdateAt:    now,
	}).Error; err != nil {
		t.Fatalf("create order item: %v", err)
	}
	if _, err := paymentService.ResolveForPurchase(
		context.Background(),
		user.ID,
		"contact_voucher_2_new",
		model.PaymentProductCodeContactVoucher,
		0,
	); !errors.Is(err, ErrPaymentPackageLimitReached) {
		t.Fatalf("expected package limit after purchase, got %v", err)
	}
}

func TestPaymentPackagePublishRequiresUniqueVirtualProductID(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.PaymentProduct{}, &model.PaymentPackage{}, &model.PaymentPackageChangeLog{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	now := time.Now()
	product := &model.PaymentProduct{
		ProductCode:   model.PaymentProductCodeContactVoucher,
		Name:          "联系券",
		SelectionMode: model.PaymentSelectionModeMultiple,
		CreateAt:      now,
		UpdateAt:      now,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	baseRepository := repository.NewRepository(nil, db)
	paymentService := NewPaymentPackageService(
		&Service{tm: repository.NewTransaction(baseRepository)},
		repository.NewPaymentPackageRepository(baseRepository),
		repository.NewPaymentProductRepository(baseRepository),
		nil,
		nil,
	)
	create := func(skuCode, virtualProductID string) int64 {
		id, err := paymentService.Create(context.Background(), PaymentPackageCreateInput{
			ProductID:        product.ID,
			SKUCode:          skuCode,
			VirtualProductID: virtualProductID,
			Name:             skuCode,
			PriceCents:       100,
			BenefitConfig:    model.PaymentBenefitConfig{ContactVouchers: 1},
			Operator:         "test",
		})
		if err != nil {
			t.Fatalf("create package %s: %v", skuCode, err)
		}
		return id
	}
	publish := func(id int64) error {
		item, err := paymentService.GetByID(context.Background(), id)
		if err != nil {
			t.Fatalf("get package %d: %v", id, err)
		}
		return paymentService.Publish(context.Background(), id, item.Package.Version, "test")
	}

	if err := publish(create("contact_voucher_without_tool", "")); !errors.Is(err, ErrVirtualPaymentUnavailable) {
		t.Fatalf("publish without virtual product ID = %v, want %v", err, ErrVirtualPaymentUnavailable)
	}
	if err := publish(create("contact_voucher_first", "wechat_tool_1")); err != nil {
		t.Fatalf("publish first package: %v", err)
	}
	if err := publish(create("contact_voucher_duplicate", "wechat_tool_1")); !errors.Is(err, ErrPaymentPackageInvalid) {
		t.Fatalf("publish duplicate virtual product ID = %v, want %v", err, ErrPaymentPackageInvalid)
	}
}
