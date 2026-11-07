package model

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestPaymentPackageAutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := db.AutoMigrate(
		&PaymentProduct{},
		&PaymentPackage{},
		&PaymentPackageChangeLog{},
	); err != nil {
		t.Fatalf("auto migrate payment package tables: %v", err)
	}

	now := time.Now()
	product := &PaymentProduct{
		ProductCode:   PaymentProductCodeContactVoucher,
		Name:          "联系券",
		SelectionMode: PaymentSelectionModeMultiple,
		CreateAt:      now,
		UpdateAt:      now,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatalf("create payment product: %v", err)
	}
	pkg := &PaymentPackage{
		ProductID:         product.ID,
		SKUCode:           "voucher_10",
		Name:              "10张联系券",
		PriceCents:        100,
		BenefitConfigJSON: `{"contact_vouchers":10}`,
		SaleRuleJSON:      `{}`,
		Status:            PaymentPackageStatusDraft,
		Version:           1,
		CreateAt:          now,
		UpdateAt:          now,
	}
	if err := db.Create(pkg).Error; err != nil {
		t.Fatalf("create payment package: %v", err)
	}
	if err := db.Create(&PaymentPackageChangeLog{
		SKUID:          pkg.ID,
		SKUVersion:     1,
		Action:         PaymentPackageActionCreate,
		BeforeSnapshot: "null",
		AfterSnapshot:  `{"sku_code":"voucher_10"}`,
		CreateAt:       now,
	}).Error; err != nil {
		t.Fatalf("create payment package change log: %v", err)
	}
}
