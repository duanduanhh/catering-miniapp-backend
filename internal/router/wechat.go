package router

import (
	"github.com/gin-gonic/gin"
)

func InitWechatRouter(deps RouterDeps, r *gin.RouterGroup) {
	noAuthRouter := r.Group("/")
	{
		noAuthRouter.POST("/wechat/user/register", deps.WechatHandler.Register)
		noAuthRouter.POST("/wechat/user/login", deps.WechatHandler.Login)
		noAuthRouter.GET("/wechat/virtual-payment/notify", deps.WechatHandler.VerifyVirtualPaymentNotify)
		noAuthRouter.POST("/wechat/virtual-payment/notify", deps.WechatHandler.VirtualPaymentNotify)
	}
}
