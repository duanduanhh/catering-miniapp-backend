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
	"github.com/go-nunu/nunu-layout-advanced/pkg/sid"
)

func TestCreateContactVoucherOrderUsesServerSKUPriceAndBenefits(t *testing.T) {
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
	user := &model.User{WechatOpenID: "openid", CreateAt: now, UpdateAt: now}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create product: %v", err)
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	baseRepository := repository.NewRepository(nil, db)
	transaction := repository.NewTransaction(baseRepository)
	baseService := &Service{tm: transaction, sid: sid.NewSid()}
	packageRepository := repository.NewPaymentPackageRepository(baseRepository)
	productRepository := repository.NewPaymentProductRepository(baseRepository)
	userRepository := repository.NewUserRepository(baseRepository)
	orderRepository := repository.NewOrderRepository(baseRepository)
	orderItemRepository := repository.NewOrderItemRepository(baseRepository)
	packageService := NewPaymentPackageService(
		baseService,
		packageRepository,
		productRepository,
		userRepository,
		orderRepository,
	)
	packageID, err := packageService.Create(context.Background(), PaymentPackageCreateInput{
		ProductID:        product.ID,
		SKUCode:          "contact_voucher_2_new",
		Name:             "2张联系券",
		PriceCents:       150,
		VirtualProductID: "contact_voucher_2_new",
		BenefitConfig:    model.PaymentBenefitConfig{ContactVouchers: 2},
		PromotionConfig: model.PaymentPromotionConfig{
			FirstPurchasePriceCents: 100,
			FirstPurchaseScope:      model.PaymentFirstPurchaseScopeProduct,
			VirtualProductID:        "contact_voucher_2_first",
		},
		Operator: "test",
	})
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	aggregate, err := packageService.GetByID(context.Background(), packageID)
	if err != nil {
		t.Fatalf("get package: %v", err)
	}
	if err := packageService.Publish(context.Background(), packageID, aggregate.Package.Version, "test"); err != nil {
		t.Fatalf("publish package: %v", err)
	}

	orderService := NewOrderService(
		baseService,
		orderRepository,
		orderItemRepository,
		repository.NewJobRepository(baseRepository),
		userRepository,
		repository.NewContactVoucherHistoryRepository(baseRepository),
		packageService,
	)
	order, item, err := orderService.CreateContactVoucherOrder(
		context.Background(),
		user.ID,
		"contact_voucher_2_new",
	)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	cents, err := order.AmountTotal.ToCents()
	if err != nil {
		t.Fatalf("read amount: %v", err)
	}
	if cents != 100 || item.ContactVoucherNum != 2 {
		t.Fatalf("unexpected order price or benefit: cents=%d item=%+v", cents, item)
	}
	if item.SKUID != packageID || item.SKUCode != "contact_voucher_2_new" ||
		item.PriceCentsSnapshot != 100 || item.BenefitSnapshot == "" || item.PromotionSnapshot == "null" {
		t.Fatalf("missing SKU snapshot: %+v", item)
	}
	// A pending or failed payment must not consume the first-purchase offer.
	next, err := packageService.ResolveForPurchase(
		context.Background(), user.ID, "contact_voucher_2_new", model.PaymentProductCodeContactVoucher, 0,
	)
	if err != nil {
		t.Fatalf("resolve first-purchase price after pending order: %v", err)
	}
	if next.PurchasePriceCents != 100 || next.Promotion == nil {
		t.Fatalf("next purchase quote = %+v, want first-purchase price", next)
	}
	if err := db.Model(&model.Order{}).Where("id = ?", order.ID).Update("status", model.OrderStatusPaid).Error; err != nil {
		t.Fatalf("mark order paid: %v", err)
	}
	next, err = packageService.ResolveForPurchase(
		context.Background(), user.ID, "contact_voucher_2_new", model.PaymentProductCodeContactVoucher, 0,
	)
	if err != nil {
		t.Fatalf("resolve regular price after paid order: %v", err)
	}
	if next.PurchasePriceCents != 150 || next.Promotion != nil {
		t.Fatalf("next purchase quote = %+v, want regular price", next)
	}
	if err := db.Model(&model.Order{}).Where("id = ?", order.ID).Update("status", model.OrderStatusCanceled).Error; err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if _, err := orderService.CompleteVirtualPaymentOrder(context.Background(), VirtualPaymentGoodsDelivery{
		OutTradeNo:       order.OrderNo,
		OpenID:           user.WechatOpenID,
		TransactionID:    "wx-transaction-1",
		ProductID:        item.VirtualProductIDSnapshot,
		Quantity:         1,
		ActualPriceCents: 100,
	}); !errors.Is(err, ErrVirtualPaymentOrderClosed) {
		t.Fatalf("closed order notification = %v, want %v", err, ErrVirtualPaymentOrderClosed)
	}
}
