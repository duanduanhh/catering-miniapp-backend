package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/pkg/wechatpay"
)

type PayService interface {
	// BuildPayParams 构建支付参数（包括调用统一下单接口）
	BuildPayParams(ctx context.Context, orderNo string, amount int64, userOpenid string, description string) (v1.PayParams, error)
	CloseOrder(ctx context.Context, orderNo string) error
}

// CloseOrder 关闭微信侧未支付订单，防止超时后继续付款。
func (s *payService) CloseOrder(ctx context.Context, orderNo string) error {
	if orderNo == "" {
		return errors.New("order_no is required")
	}
	if err := s.wechatPayClient.CloseOrder(ctx, orderNo); err != nil {
		return fmt.Errorf("close wechat order: %w", err)
	}
	return nil
}

type payService struct {
	config          *viper.Viper
	wechatPayClient *wechatpay.WechatPayClient
	logger          *zap.Logger
}

func NewPayService(config *viper.Viper, logger *zap.Logger) (PayService, error) {
	// 创建微信支付客户端
	wechatPayClient, err := wechatpay.NewWechatPayClient(config)
	if err != nil {
		return nil, fmt.Errorf("create wechat pay client failed: %w", err)
	}

	return &payService{
		config:          config,
		wechatPayClient: wechatPayClient,
		logger:          logger,
	}, nil
}

// BuildPayParams 构建支付参数
// 流程：
// 1. 调用微信统一下单接口获取 prepay_id
// 2. 使用 prepay_id 生成支付参数（包括签名）
// 3. 返回给前端供 wx.requestPayment() 调用
func (s *payService) BuildPayParams(ctx context.Context, orderNo string, amount int64, userOpenid string, description string) (v1.PayParams, error) {
	// 验证参数
	if orderNo == "" {
		return v1.PayParams{}, errors.New("order_no is required")
	}
	if amount <= 0 {
		return v1.PayParams{}, errors.New("amount must be greater than 0")
	}
	if userOpenid == "" {
		return v1.PayParams{}, errors.New("user_openid is required")
	}
	if description == "" {
		description = "订单支付"
	}

	// 获取通知URL
	notifyURL := s.config.GetString("wxpay.notify_url")
	if notifyURL == "" {
		return v1.PayParams{}, errors.New("wxpay.notify_url is not configured")
	}

	// 调用微信统一下单接口
	prepayRequest := &wechatpay.PreorderRequest{
		Description: description,
		OutTradeNo:  orderNo,
		Amount:      amount, // 单位为分
		Openid:      userOpenid,
		NotifyURL:   notifyURL,
	}

	prepayID, err := s.wechatPayClient.PrepayWithCode(ctx, prepayRequest)
	if err != nil {
		s.logger.Error("PrepayWithCode failed", zap.Error(err), zap.String("order_no", orderNo))
		return v1.PayParams{}, fmt.Errorf("prepay failed: %w", err)
	}

	s.logger.Info("PrepayWithCode success", zap.String("order_no", orderNo), zap.String("prepay_id", prepayID))

	// 生成支付参数
	payParams, err := s.wechatPayClient.GeneratePayParams(ctx, prepayID)
	if err != nil {
		s.logger.Error("GeneratePayParams failed", zap.Error(err), zap.String("prepay_id", prepayID))
		return v1.PayParams{}, fmt.Errorf("generate pay params failed: %w", err)
	}

	// 转换为 API 响应格式
	return v1.PayParams{
		TimeStamp: payParams.TimeStamp,
		NonceStr:  payParams.NonceStr,
		Package:   payParams.Package,
		SignType:  payParams.SignType,
		PaySign:   payParams.PaySign,
	}, nil
}
