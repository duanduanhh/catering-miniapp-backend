package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/spf13/viper"
)

func TestVirtualPaymentPrepare(t *testing.T) {
	config := viper.New()
	config.Set("virtual_payment.offer_id", "offer_test")
	config.Set("virtual_payment.sandbox_app_key", "app_key_test")
	config.Set("virtual_payment.env", 1)
	service := NewVirtualPaymentService(config)

	params, err := service.Prepare(context.Background(), &VirtualPaymentOrder{
		OrderNo:          "TOP202608010001",
		AmountCents:      390,
		VirtualProductID: "wx_job_top_1d",
	}, "session_key_test")
	if err != nil {
		t.Fatalf("Prepare() error = %v", err)
	}
	if params.Mode != virtualPaymentModeGoods {
		t.Fatalf("mode = %q, want %q", params.Mode, virtualPaymentModeGoods)
	}
	var signData virtualPaymentSignData
	if err := json.Unmarshal([]byte(params.SignData), &signData); err != nil {
		t.Fatalf("signData is not JSON: %v", err)
	}
	if signData.OfferID != "offer_test" || signData.ProductID != "wx_job_top_1d" || signData.GoodsPrice != 390 || signData.OutTradeNo != "TOP202608010001" || signData.Env != 1 {
		t.Fatalf("unexpected signData: %+v", signData)
	}
	if params.PaySig != hmacSHA256Hex("app_key_test", virtualPaymentModeGoods+"&"+params.SignData) {
		t.Fatal("paySig does not match the official virtual payment signing payload")
	}
	if params.Signature != hmacSHA256Hex("session_key_test", params.SignData) {
		t.Fatal("signature does not match the session-key signing payload")
	}
}

func TestVirtualPaymentPrepareRequiresConfiguration(t *testing.T) {
	_, err := NewVirtualPaymentService(viper.New()).Prepare(context.Background(), &VirtualPaymentOrder{
		OrderNo: "TOP202608010001", AmountCents: 390, VirtualProductID: "wx_job_top_1d",
	}, "session_key_test")
	if err != ErrVirtualPaymentUnavailable {
		t.Fatalf("Prepare() error = %v, want %v", err, ErrVirtualPaymentUnavailable)
	}
}
