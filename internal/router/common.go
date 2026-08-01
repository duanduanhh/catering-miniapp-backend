package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitCommonRouter(deps RouterDeps, r *gin.RouterGroup) {
	// 通用接口，无需认证
	commonRouter := r.Group("/")
	{
		commonRouter.GET("/close_reasons", deps.JobHandler.GetCloseReasons)
	}
	packageRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger, deps.UserRepo))
	packageRouter.POST("/payment-packages/list", deps.PaymentPackageHandler.ListAvailable)
}
