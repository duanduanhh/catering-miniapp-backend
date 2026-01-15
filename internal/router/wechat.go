package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitWechatRouter(deps RouterDeps, r *gin.RouterGroup) {
	noAuthRouter := r.Group("/")
	{
		noAuthRouter.POST("/wechat/user/register", deps.WechatHandler.Register)
		noAuthRouter.POST("/wechat/user/login", deps.WechatHandler.Login)
		noAuthRouter.POST("/wechat/pay/notify", deps.WechatHandler.PayNotify)
	}
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/wechat/pay", deps.WechatHandler.Pay)
	}
}
