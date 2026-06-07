package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
	"github.com/go-nunu/nunu-layout-advanced/pkg/wechatpay"
)

type WechatHandler struct {
	*Handler
	orderService    service.OrderService
	wechatService   service.WechatService
	wechatPayClient *wechatpay.WechatPayClient
}

func NewWechatHandler(handler *Handler, orderService service.OrderService, wechatService service.WechatService, wechatPayClient *wechatpay.WechatPayClient) *WechatHandler {
	return &WechatHandler{
		Handler:         handler,
		orderService:    orderService,
		wechatService:   wechatService,
		wechatPayClient: wechatPayClient,
	}
}

// Register godoc
// @Summary 微信注册
// @Description 通过微信小程序注册新用户，需同时传手机号授权 code 和登录 code。inviter_id 为可选邀请人用户ID。注册成功后不返回 token，需再调用 /login 获取。若 openid 已注册返回 400。
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body v1.WechatRegisterRequest true "params"
// @Success 200 {object} v1.WechatLoginResponseData
// @Router /wechat/user/register [post]
func (h *WechatHandler) Register(ctx *gin.Context) {
	var req v1.WechatRegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	_, user, err := h.wechatService.Register(ctx, req.PhoneCode, req.LoginCode, req.InviterID)
	if err != nil {
		h.logger.WithContext(ctx).Error("wechatService.Register error", zap.Error(err))
		if err == service.ErrUserExists {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.WechatLoginResponseData{
		UserInfo: v1.WechatLoginUserInfo{ID: user.ID},
	})
}

// Login godoc
// @Summary 微信登录
// @Description 通过微信小程序 login code 换取 JWT token。用户不存在时返回 404，需先调用 /register。token 放在后续请求的 `token` header 中（小写，非 Authorization）。
// @Tags 用户模块
// @Accept json
// @Produce json
// @Param request body v1.WechatLoginRequest true "params"
// @Success 200 {object} v1.WechatLoginResponseData
// @Router /wechat/user/login [post]
func (h *WechatHandler) Login(ctx *gin.Context) {
	var req v1.WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}
	token, user, isOldUser, err := h.wechatService.Login(ctx, req.LoginCode)
	if err != nil {
		h.logger.WithContext(ctx).Error("wechatService.Login error", zap.Error(err))
		if err == service.ErrUserNotFound {
			v1.HandleError(ctx, http.StatusNotFound, v1.ErrNotFound, err.Error())
			return
		}
		v1.HandleError(ctx, http.StatusInternalServerError, v1.ErrInternalServerError, err.Error())
		return
	}
	v1.HandleSuccess(ctx, v1.WechatLoginResponseData{
		Token:     token,
		IsOldUser: isOldUser,
		UserInfo:  v1.WechatLoginUserInfo{ID: user.ID},
	})
}

// PayNotify godoc
// @Summary 微信支付回调（仅供微信服务器调用）
// @Description 微信支付结果异步通知接口，由微信服务器主动调用，前端无需关注。收到回调后解密验签，金额一致则更新订单状态并执行对应业务逻辑（置顶/券充值/刷新）。
// @Tags 支付模块
// @Accept json
// @Produce json
// @Param request body v1.WechatPayNotifyRequest true "params"
// @Success 200 {object} v1.Response
// @Router /wechat/pay/notify [post]
func (h *WechatHandler) PayNotify(ctx *gin.Context) {
	// 读取原始请求体
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		h.logger.WithContext(ctx).Error("read request body error", zap.Error(err))
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, "invalid request body")
		return
	}

	// 解密通知数据
	decryptedData, err := h.wechatPayClient.DecryptPaymentNotify(body)
	if err != nil {
		h.logger.WithContext(ctx).Error("DecryptPaymentNotify error", zap.Error(err), zap.ByteString("body", body))
		v1.HandleError(ctx, http.StatusBadRequest, v1.ErrBadRequest, err.Error())
		return
	}

	// 验证交易状态
	if decryptedData.TradeState != "SUCCESS" {
		h.logger.WithContext(ctx).Info("trade state not success", zap.String("trade_state", decryptedData.TradeState), zap.String("out_trade_no", decryptedData.OutTradeNo))
		// 返回成功以免微信重复回调
		v1.HandleSuccess(ctx, nil)
		return
	}

	// 金额单位转换：分 -> 元（数据库存储为 float64）
	var amount float64
	if decryptedData.Amount != nil {
		amount = float64(decryptedData.Amount.Total) / 100.0
	}

	// 调用 service 处理订单支付
	_, err = h.orderService.PayOrderByNotify(ctx, decryptedData.OutTradeNo, amount, "wechat", decryptedData.TransactionId)
	if err != nil {
		h.logger.WithContext(ctx).Error("orderService.PayOrderByNotify error", zap.Error(err), zap.String("order_no", decryptedData.OutTradeNo))
		if err == service.ErrAmountMismatch {
			v1.HandleError(ctx, http.StatusBadRequest, v1.ErrAmountMismatch, err.Error())
			return
		}
		// 即使处理失败也返回成功，避免微信重复回调
		v1.HandleSuccess(ctx, nil)
		return
	}

	h.logger.WithContext(ctx).Info("pay notify success", zap.String("order_no", decryptedData.OutTradeNo), zap.Int64("amount", decryptedData.Amount.Total))

	v1.HandleSuccess(ctx, nil)
}
