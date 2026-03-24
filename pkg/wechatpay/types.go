package wechatpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

// PreorderRequest 预下单请求
type PreorderRequest struct {
	Description string  // 商品描述
	OutTradeNo  string  // 商户订单号
	Amount      int64   // 金额（分）
	Openid      string  // 用户OpenID
	NotifyURL   string  // 通知URL
	TimeExpire  *string // 交易结束时间（可选）
}

// PayParams 支付参数（返回给前端）
type PayParams struct {
	AppID     string `json:"appId"`     // 应用ID
	TimeStamp string `json:"timeStamp"` // 时间戳
	NonceStr  string `json:"nonceStr"`  // 随机字符串
	Package   string `json:"package"`   // 数据包
	SignType  string `json:"signType"`  // 签名类型
	PaySign   string `json:"paySign"`   // 签名
	prepayID  string // 内部字段，不导出
}

// SetPrepayID 设置 prepay_id
func (p *PayParams) SetPrepayID(prepayID string) {
	p.prepayID = prepayID
}

// Sign 对支付参数进行签名
func (p *PayParams) Sign(privateKey *rsa.PrivateKey) error {
	// 生成随机字符串
	p.NonceStr = generateNonceStr(32)
	p.TimeStamp = strconv.FormatInt(time.Now().Unix(), 10)
	p.Package = fmt.Sprintf("prepay_id=%s", p.prepayID)
	p.SignType = "RSA"

	// 生成签名
	payload := fmt.Sprintf("%s\n%s\n%s\n%s\n",
		p.AppID,
		p.TimeStamp,
		p.NonceStr,
		p.Package,
	)

	// 使用 SHA256 + RSA 签名
	hash := sha256.Sum256([]byte(payload))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hash[:])
	if err != nil {
		return fmt.Errorf("sign failed: %w", err)
	}

	p.PaySign = base64.StdEncoding.EncodeToString(signature)
	return nil
}

// generateNonceStr 生成随机字符串
func generateNonceStr(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
