package handler

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVirtualPaymentCallbackSignature(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest("POST", "/wechat/virtual-payment/notify?signature=plain&msg_signature=encrypted", nil)
	if got := virtualPaymentCallbackSignature(ctx); got != "encrypted" {
		t.Fatalf("signature = %q, want msg_signature", got)
	}

	plainCtx, _ := gin.CreateTestContext(httptest.NewRecorder())
	plainCtx.Request = httptest.NewRequest("POST", "/wechat/virtual-payment/notify?signature=plain", nil)
	if got := virtualPaymentCallbackSignature(plainCtx); got != "plain" {
		t.Fatalf("signature = %q, want signature", got)
	}
}
