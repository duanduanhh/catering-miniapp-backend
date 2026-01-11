package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitCollectRouter(deps RouterDeps, r *gin.RouterGroup) {
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/jobs/collect", deps.CollectHandler.Collect)
		strictAuthRouter.POST("/jobs/cancnel_collect", deps.CollectHandler.Cancel)
		strictAuthRouter.POST("/collect/my", deps.CollectHandler.My)
	}
}
