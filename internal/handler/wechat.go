package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	v1 "github.com/go-nunu/nunu-layout-advanced/api/v1"
	"github.com/go-nunu/nunu-layout-advanced/internal/service"
)

type WechatHandler struct {
	*Handler
	orderService                service.OrderService
	wechatService               service.WechatService
	virtualPaymentNotifyService service.VirtualPaymentNotifyService
}

func NewWechatHandler(
	handler *Handler,
	orderService service.OrderService,
	wechatService service.WechatService,
	virtualPaymentNotifyService service.VirtualPaymentNotifyService,
) *WechatHandler {
	return &WechatHandler{
		Handler:                     handler,
		orderService:                orderService,
		wechatService:               wechatService,
		virtualPaymentNotifyService: virtualPaymentNotifyService,
	}
}

// Register godoc
// @Summary 微信注册
// @Description 通过微信小程序注册新用户，需同时传手机号授权 code 和登录 code。inviter_id 为可选邀请人用户ID。注册成功后直接返回 token，无需再调 /login。若 openid 已注册返回 400。
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
	token, user, err := h.wechatService.Register(ctx, req.PhoneCode, req.LoginCode, req.InviterID)
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
		Token:    token,
		UserInfo: v1.WechatLoginUserInfo{ID: user.ID, UserCode: user.UserCode},
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
		UserInfo:  v1.WechatLoginUserInfo{ID: user.ID, UserCode: user.UserCode},
	})
}

// VerifyVirtualPaymentNotify godoc
// @Summary 验证虚拟支付消息推送地址（仅供微信服务器调用）
// @Description 在微信虚拟支付后台保存发货推送地址时调用。接口校验 message_token 签名；若后台选择安全模式，同时解密 echostr。
// @Tags 支付模块
// @Produce text/plain
// @Param signature query string false "明文模式微信签名"
// @Param msg_signature query string false "安全模式微信签名"
// @Param timestamp query string true "时间戳"
// @Param nonce query string true "随机数"
// @Param echostr query string true "微信回显内容"
// @Success 200 {string} string "微信回显内容"
// @Router /wechat/virtual-payment/notify [get]
func (h *WechatHandler) VerifyVirtualPaymentNotify(ctx *gin.Context) {
	echo, err := h.virtualPaymentNotifyService.VerifyURL(
		virtualPaymentCallbackSignature(ctx), ctx.Query("timestamp"), ctx.Query("nonce"), ctx.Query("echostr"),
	)
	if err != nil {
		h.logger.WithContext(ctx).Warn("verify virtual payment callback failed", zap.Error(err))
		ctx.Status(http.StatusForbidden)
		return
	}
	ctx.String(http.StatusOK, echo)
}

// VirtualPaymentNotify godoc
// @Summary 微信虚拟支付道具发货推送（仅供微信服务器调用）
// @Description 接收 xpay_goods_deliver_notify。服务端校验微信消息签名（安全模式下解密 XML），原子更新订单为已支付，并幂等发放订单权益；招租订单会自动发布对应招租信息。返回 {"ErrCode":0} 表示发货完成。
// @Tags 支付模块
// @Accept xml
// @Produce json
// @Param signature query string false "明文模式微信签名"
// @Param msg_signature query string false "安全模式微信签名"
// @Success 200 {object} map[string]interface{} "发货成功"
// @Router /wechat/virtual-payment/notify [post]
func (h *WechatHandler) VirtualPaymentNotify(ctx *gin.Context) {
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		virtualPaymentNotifyError(ctx, http.StatusBadRequest, "invalid request body")
		return
	}
	notice, err := h.virtualPaymentNotifyService.Parse(
		virtualPaymentCallbackSignature(ctx), ctx.Query("timestamp"), ctx.Query("nonce"), body,
	)
	if err != nil {
		h.logger.WithContext(ctx).Warn("parse virtual payment notification failed", zap.Error(err))
		if err == service.ErrVirtualPaymentNotifySignature || err == service.ErrVirtualPaymentNotifyUnavailable {
			virtualPaymentNotifyError(ctx, http.StatusForbidden, "invalid notification")
			return
		}
		virtualPaymentNotifyError(ctx, http.StatusBadRequest, "invalid notification")
		return
	}

	if _, err = h.orderService.CompleteVirtualPaymentOrder(ctx, *notice); err != nil {
		h.logger.WithContext(ctx).Error("process virtual payment notification failed", zap.Error(err), zap.String("order_no", notice.OutTradeNo))
		// 非 2xx 会触发微信重试，直到订单成功处理或达到平台重试上限。
		virtualPaymentNotifyError(ctx, http.StatusInternalServerError, "delivery failed")
		return
	}
	h.logger.WithContext(ctx).Info("virtual payment delivery completed", zap.String("order_no", notice.OutTradeNo), zap.String("transaction_id", notice.TransactionID))
	ctx.JSON(http.StatusOK, gin.H{"ErrCode": 0, "ErrMsg": "success"})
}

// virtualPaymentCallbackSignature 兼容微信两种消息推送模式：
// 明文模式使用 signature，安全模式使用 msg_signature。
func virtualPaymentCallbackSignature(ctx *gin.Context) string {
	if signature := ctx.Query("msg_signature"); signature != "" {
		return signature
	}
	return ctx.Query("signature")
}

func virtualPaymentNotifyError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"ErrCode": -1, "ErrMsg": message})
}
