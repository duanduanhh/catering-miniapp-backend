package wechatpay

import (
	"encoding/json"
	"time"
)

// PaymentNotifyData 支付通知数据结构
type PaymentNotifyData struct {
	ID           string        `json:"id"`            // 通知ID
	CreateTime   *time.Time    `json:"create_time"`   // 通知创建时间
	EventType    string        `json:"event_type"`    // 事件类型
	ResourceType string        `json:"resource_type"` // 资源类型
	Summary      string        `json:"summary"`       // 摘要
	Resource     *ResourceData `json:"resource"`      // 资源数据
}

// ResourceData 资源数据
type ResourceData struct {
	Algorithm      string `json:"algorithm"`       // 加密算法
	Ciphertext     string `json:"ciphertext"`      // 加密数据
	OriginalType   string `json:"original_type"`   // 原始类型
	AssociatedData string `json:"associated_data"` // 关联数据
	Nonce          string `json:"nonce"`           // 随机数
}

// DecryptedData 解密后的数据
type DecryptedData struct {
	Appid          string      `json:"appid"`            // 应用ID
	Mchid          string      `json:"mchid"`            // 商户ID
	OutTradeNo     string      `json:"out_trade_no"`     // 商户订单号
	TransactionId  string      `json:"transaction_id"`   // 微信支付订单号
	TradeType      string      `json:"trade_type"`       // 交易类型
	TradeState     string      `json:"trade_state"`      // 交易状态
	TradeStateDesc string      `json:"trade_state_desc"` // 交易状态描述
	SuccessTime    *time.Time  `json:"success_time"`     // 支付成功时间
	Payer          *PayerInfo  `json:"payer"`            // 支付者信息
	Amount         *AmountInfo `json:"amount"`           // 金额信息
	SceneInfo      *SceneInfo  `json:"scene_info"`       // 场景信息
}

// PayerInfo 支付者信息
type PayerInfo struct {
	Openid string `json:"openid"` // 用户OpenID
}

// AmountInfo 金额信息
type AmountInfo struct {
	Total         int64  `json:"total"`          // 总金额（单位：分）
	PayerTotal    int64  `json:"payer_total"`    // 支付者支付金额（单位：分）
	Currency      string `json:"currency"`       // 货币类型
	PayerCurrency string `json:"payer_currency"` // 支付者货币类型
}

// SceneInfo 场景信息
type SceneInfo struct {
	DeviceID string `json:"device_id"` // 设备ID
}

// PaymentSuccessCallback 支付成功回调
type PaymentSuccessCallback struct {
	OutTradeNo    string     // 商户订单号
	TransactionId string     // 微信支付订单号
	Amount        int64      // 金额（分）
	SuccessTime   *time.Time // 支付成功时间
}

// ParsePaymentNotify 解析支付通知
func ParsePaymentNotify(data []byte) (*PaymentNotifyData, error) {
	var notify PaymentNotifyData
	if err := json.Unmarshal(data, &notify); err != nil {
		return nil, err
	}
	return &notify, nil
}

// ExtractDecryptedData 提取解密数据
func ExtractDecryptedData(data []byte) (*DecryptedData, error) {
	var decrypted DecryptedData
	if err := json.Unmarshal(data, &decrypted); err != nil {
		return nil, err
	}
	return &decrypted, nil
}
