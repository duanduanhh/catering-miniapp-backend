package wechatpay

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

// WechatPayClient 微信支付客户端
type WechatPayClient struct {
	client     *core.Client
	appID      string
	mchID      string
	privateKey *rsa.PrivateKey
	apiV3Key   string
	notifyURL  string
}

// NewWechatPayClient 创建微信支付客户端
func NewWechatPayClient(config *viper.Viper) (*WechatPayClient, error) {
	appID := config.GetString("wxpay.app_id")
	mchID := config.GetString("wxpay.mch_id")
	certPath := config.GetString("wxpay.cert_path")
	serialNumber := config.GetString("wxpay.serial_number")
	apiV3Key := config.GetString("wxpay.api_v3_key")
	notifyURL := config.GetString("wxpay.notify_url")

	if appID == "" || mchID == "" || certPath == "" || serialNumber == "" || apiV3Key == "" {
		return nil, fmt.Errorf("missing required wxpay configuration")
	}

	// 使用官方函数从文件加载商户私钥
	privateKey, err := utils.LoadPrivateKeyWithPath(certPath)
	if err != nil {
		return nil, fmt.Errorf("load private key failed: %w", err)
	}

	// 创建微信支付客户端，并自动获取平台证书
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, serialNumber, privateKey, apiV3Key),
	}

	client, err := core.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("create wechat pay client failed: %w", err)
	}

	return &WechatPayClient{
		client:     client,
		appID:      appID,
		mchID:      mchID,
		privateKey: privateKey,
		apiV3Key:   apiV3Key,
		notifyURL:  notifyURL,
	}, nil
}

// PrepayWithCode 调用统一下单接口获取 prepay_id
// 直接使用官方库 JsapiApiService.Prepay 方法
func (c *WechatPayClient) PrepayWithCode(ctx context.Context, req *PreorderRequest) (string, error) {
	if req.NotifyURL == "" {
		req.NotifyURL = c.notifyURL
	}

	notifyURL := req.NotifyURL
	if notifyURL == "" {
		return "", fmt.Errorf("notify_url is required")
	}

	// 使用官方库的 JsapiApiService 进行下单
	svc := (*jsapi.JsapiApiService)(&services.Service{Client: c.client})

	prepayReq := jsapi.PrepayRequest{
		Appid:       core.String(c.appID),
		Mchid:       core.String(c.mchID),
		Description: core.String(req.Description),
		OutTradeNo:  core.String(req.OutTradeNo),
		NotifyUrl:   core.String(notifyURL),
		Amount: &jsapi.Amount{
			Total:    core.Int64(req.Amount), // 单位为分
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(req.Openid),
		},
	}

	// 调用官方 Prepay 方法
	resp, result, err := svc.Prepay(ctx, prepayReq)
	if err != nil {
		log.Printf("Prepay failed: %v, result: %v", err, result)
		return "", fmt.Errorf("prepay failed: %w", err)
	}

	// 检查响应
	if result != nil && result.Response != nil {
		log.Printf("Prepay response status: %d", result.Response.StatusCode)
	}

	if resp == nil || resp.PrepayId == nil {
		return "", fmt.Errorf("prepay response is empty or prepay_id is nil")
	}

	prepayID := *resp.PrepayId
	log.Printf("PrepayWithCode success: prepay_id=%s, out_trade_no=%s", prepayID, req.OutTradeNo)
	return prepayID, nil
}

// CloseOrder 关闭 JSAPI 未支付订单，关闭成功后微信不再允许用户继续付款。
func (c *WechatPayClient) CloseOrder(ctx context.Context, outTradeNo string) error {
	if outTradeNo == "" {
		return errors.New("out_trade_no is required")
	}
	svc := (*jsapi.JsapiApiService)(&services.Service{Client: c.client})
	_, err := svc.CloseOrder(ctx, jsapi.CloseOrderRequest{
		OutTradeNo: core.String(outTradeNo),
		Mchid:      core.String(c.mchID),
	})
	if err != nil {
		return fmt.Errorf("close order failed: %w", err)
	}
	return nil
}

// GeneratePayParams 生成支付参数（包括签名）
func (c *WechatPayClient) GeneratePayParams(ctx context.Context, prepayID string) (*PayParams, error) {
	payParams := &PayParams{
		AppID: c.appID,
	}
	payParams.SetPrepayID(prepayID)

	err := payParams.Sign(c.privateKey)
	if err != nil {
		return nil, fmt.Errorf("sign pay params failed: %w", err)
	}

	return payParams, nil
}

// GetAppID 获取应用ID
func (c *WechatPayClient) GetAppID() string {
	return c.appID
}

// GetMchID 获取商户ID
func (c *WechatPayClient) GetMchID() string {
	return c.mchID
}

// GetAPIV3Key 获取 API v3 密钥（用于解密回调数据）
func (c *WechatPayClient) GetAPIV3Key() string {
	return c.apiV3Key
}

// DecryptNotifyResource 解密通知资源数据
// ciphertext: Base64 编码的加密数据
// nonce: 随机字符串
// associatedData: 关联数据
func (c *WechatPayClient) DecryptNotifyResource(ciphertext, nonce, associatedData string) ([]byte, error) {
	if c.apiV3Key == "" {
		return nil, errors.New("api v3 key is not configured")
	}

	// Base64 解码
	cipherBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return nil, errors.New("failed to decode ciphertext: " + err.Error())
	}

	// 创建 AES GCM  cipher
	block, err := aes.NewCipher([]byte(c.apiV3Key))
	if err != nil {
		return nil, errors.New("failed to create cipher: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("failed to create gcm: " + err.Error())
	}

	// 解密
	plaintext, err := gcm.Open(nil, []byte(nonce), cipherBytes, []byte(associatedData))
	if err != nil {
		return nil, errors.New("failed to decrypt: " + err.Error())
	}

	return plaintext, nil
}

// DecryptPaymentNotify 解密支付通知
// data: 通知请求体 JSON
// 返回解密后的数据
func (c *WechatPayClient) DecryptPaymentNotify(data []byte) (*DecryptedData, error) {
	// 解析通知数据
	var notify PaymentNotifyData
	if err := json.Unmarshal(data, &notify); err != nil {
		return nil, errors.New("failed to parse notification: " + err.Error())
	}

	// 检查事件类型
	if notify.EventType != "TRANSACTION.SUCCESS" {
		return nil, errors.New("event type is not TRANSACTION.SUCCESS: " + notify.EventType)
	}

	// 检查资源数据
	if notify.Resource == nil {
		return nil, errors.New("resource is nil")
	}

	// 解密
	plaintext, err := c.DecryptNotifyResource(
		notify.Resource.Ciphertext,
		notify.Resource.Nonce,
		notify.Resource.AssociatedData,
	)
	if err != nil {
		return nil, errors.New("failed to decrypt resource: " + err.Error())
	}

	// 解析解密后的数据
	var decrypted DecryptedData
	if err := json.Unmarshal(plaintext, &decrypted); err != nil {
		return nil, errors.New("failed to parse decrypted data: " + err.Error())
	}

	return &decrypted, nil
}
