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
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.GET("/user/info", deps.UserHandler.GetInfo)
		strictAuthRouter.POST("/user/update/geo", deps.UserHandler.UpdateGeo)
		strictAuthRouter.POST("/user/update/info", deps.UserHandler.UpdateInfo)
	}
}
