package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitWechatRouter(deps RouterDeps, r *gin.RouterGroup) {
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/wechat/pay", deps.WechatHandler.Pay)
	}
}
