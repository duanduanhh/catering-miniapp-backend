package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitJobRouter(deps RouterDeps, r *gin.RouterGroup) {
	noAuthRouter := r.Group("/")
	{
		noAuthRouter.POST("/jobs/list", deps.JobHandler.List)
		noAuthRouter.POST("/jobs/info", deps.JobHandler.Info)
	}

	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/jobs/create", deps.JobHandler.Create)
		strictAuthRouter.POST("/jobs/update", deps.JobHandler.Update)
		strictAuthRouter.POST("/jobs/refresh", deps.JobHandler.Refresh)
		strictAuthRouter.POST("/jobs/refresh/pay", deps.JobHandler.RefreshPay)
		strictAuthRouter.POST("/jobs/close", deps.JobHandler.Close)
		strictAuthRouter.POST("/jobs/my", deps.JobHandler.My)
		strictAuthRouter.POST("/jobs/top", deps.JobHandler.Top)
	}
}
