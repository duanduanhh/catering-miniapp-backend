package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitVoucherRouter(deps RouterDeps, r *gin.RouterGroup) {
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/contact_voucher/buy", deps.ContactVoucherHistoryHandler.Buy)
		strictAuthRouter.POST("/contact_voucher/cost", deps.ContactVoucherHistoryHandler.Cost)
		strictAuthRouter.POST("/contact_voucher/records", deps.ContactVoucherHistoryHandler.Records)
	}
}
