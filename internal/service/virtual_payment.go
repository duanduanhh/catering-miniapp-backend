package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"

	"github.com/spf13/viper"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
)

const (
	virtualPaymentModeGoods = "short_series_goods"
	// virtualPaymentPaySigURI 是微信虚拟支付协议规定的固定签名 URI，
	// 不能使用支付模式 short_series_goods 替代。
	virtualPaymentPaySigURI = "requestVirtualPayment"
)

type VirtualPaymentService interface {
	Prepare(ctx context.Context, order *VirtualPaymentOrder, sessionKey string) (v1.VirtualPaymentParams, error)
}

type virtualPaymentService struct {
	config *viper.Viper
}

func NewVirtualPaymentService(config *viper.Viper) VirtualPaymentService {
	return &virtualPaymentService{config: config}
}

type virtualPaymentSignData struct {
	OfferID      string `json:"offerId"`
	BuyQuantity  int    `json:"buyQuantity"`
	Env          int    `json:"env"`
	CurrencyType string `json:"currencyType"`
	ProductID    string `json:"productId"`
	GoodsPrice   int64  `json:"goodsPrice"`
	OutTradeNo   string `json:"outTradeNo"`
	Attach       string `json:"attach"`
}

func (s *virtualPaymentService) Prepare(_ context.Context, order *VirtualPaymentOrder, sessionKey string) (v1.VirtualPaymentParams, error) {
	if order == nil || order.OrderNo == "" || order.AmountCents <= 0 || order.VirtualProductID == "" || sessionKey == "" {
		return v1.VirtualPaymentParams{}, ErrVirtualPaymentUnavailable
	}
	offerID := strings.TrimSpace(s.config.GetString("virtual_payment.offer_id"))
	env := s.config.GetInt("virtual_payment.env")
	appKeyName := "virtual_payment.app_key"
	if env == 1 {
		appKeyName = "virtual_payment.sandbox_app_key"
	}
	appKey := strings.TrimSpace(s.config.GetString(appKeyName))
	if offerID == "" || appKey == "" {
		return v1.VirtualPaymentParams{}, ErrVirtualPaymentUnavailable
	}
	signData, err := json.Marshal(virtualPaymentSignData{
		OfferID:      offerID,
		BuyQuantity:  1,
		Env:          env,
		CurrencyType: "CNY",
		ProductID:    order.VirtualProductID,
		GoodsPrice:   order.AmountCents,
		OutTradeNo:   order.OrderNo,
		Attach:       order.OrderNo,
	})
	if err != nil {
		return v1.VirtualPaymentParams{}, errors.New("marshal virtual payment sign data")
	}
	raw := string(signData)
	return v1.VirtualPaymentParams{
		SignData:  raw,
		PaySig:    hmacSHA256Hex(appKey, virtualPaymentPaySigURI+"&"+raw),
		Signature: hmacSHA256Hex(sessionKey, raw),
		Mode:      virtualPaymentModeGoods,
	}, nil
}

func hmacSHA256Hex(key, payload string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
