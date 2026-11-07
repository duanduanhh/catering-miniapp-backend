package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitUserRouter(
	deps RouterDeps,
	r *gin.RouterGroup,
) {
	// No route group has permission
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger, deps.UserRepo))
	{
		strictAuthRouter.GET("/user/info", deps.UserHandler.GetInfo)
		strictAuthRouter.POST("/user/update/geo", deps.UserHandler.UpdateGeo)
		strictAuthRouter.POST("/user/update/info", deps.UserHandler.UpdateInfo)
		strictAuthRouter.POST("/user/orders", deps.OrderHandler.ListOrders)
		strictAuthRouter.POST("/order/status", deps.OrderHandler.QueryOrderStatus)
		strictAuthRouter.POST("/payment/virtual/prepare", deps.OrderHandler.PrepareVirtualPayment)
		strictAuthRouter.POST("/user/invites", deps.UserHandler.ListInvites)
	}
}
