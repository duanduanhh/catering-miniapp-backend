package handler

import (
	"testing"

	"github.com/go-nunu/nunu-layout-advanced/internal/model"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

func TestPaymentPackageItemIncludesSaleRuleForAdmin(t *testing.T) {
	aggregate := service.PaymentPackageAggregate{
		Product: &model.PaymentProduct{ProductCode: model.PaymentProductCodeJobTop},
		Package: &model.PaymentPackage{ID: 1, ProductID: 2, SKUCode: "job_top_1d_new"},
		SaleRule: model.PaymentSaleRule{
			MaxPurchasePerUser: 1,
		},
	}

	adminItem := toPaymentPackageItem(aggregate, true)
	if adminItem.SaleRule.MaxPurchasePerUser != 1 {
		t.Fatalf("admin sale_rule = %+v, want limit 1", adminItem.SaleRule)
	}

	nonAdminItem := toPaymentPackageItem(aggregate, false)
	if nonAdminItem.SaleRule.MaxPurchasePerUser != 0 {
		t.Fatalf("non-admin item unexpectedly exposes sale_rule: %+v", nonAdminItem.SaleRule)
	}
}
