package server

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
)

func TestSeedPaymentProductsKeepsConfiguredPurchaseNotice(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.PaymentProduct{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	product := &model.PaymentProduct{
		ProductCode:    model.PaymentProductCodePaidRefresh,
		Name:           "付费刷新",
		SelectionMode:  model.PaymentSelectionModeSingle,
		PurchaseNotice: "后台自定义须知",
		CreateAt:       time.Now(),
		UpdateAt:       time.Now(),
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create configured product: %v", err)
	}
	if err := (&MigrateServer{db: db}).seedPaymentProducts(); err != nil {
		t.Fatalf("seed payment products: %v", err)
	}
	var refreshed model.PaymentProduct
	if err := db.First(&refreshed, product.ID).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if refreshed.PurchaseNotice != product.PurchaseNotice {
		t.Fatalf("configured purchase notice was overwritten: %q", refreshed.PurchaseNotice)
	}
}

func TestPaymentProductMigrationRejectsDuplicateSingleSKUs(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&model.PaymentProduct{}, &model.PaymentPackage{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	migration := &MigrateServer{db: db}
	if err := migration.seedPaymentProducts(); err != nil {
		t.Fatalf("seed payment products: %v", err)
	}
	var product model.PaymentProduct
	if err := db.Where("product_code = ?", model.PaymentProductCodePaidRefresh).First(&product).Error; err != nil {
		t.Fatalf("load product: %v", err)
	}
	now := time.Now()
	for _, skuCode := range []string{"paid_refresh_1", "paid_refresh_2"} {
		if err := db.Create(&model.PaymentPackage{ProductID: product.ID, SKUCode: skuCode, Name: skuCode, PriceCents: 100, Status: model.PaymentPackageStatusDraft, Version: 1, CreateAt: now, UpdateAt: now}).Error; err != nil {
			t.Fatalf("create SKU: %v", err)
		}
	}
	if err := migration.validateSingleProductCardinality(); err == nil {
		t.Fatal("expected duplicate single SKUs to be rejected")
	}
}
