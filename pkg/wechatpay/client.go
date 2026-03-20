package wechatpay

import (
	"context"
	"crypto/rsa"
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
