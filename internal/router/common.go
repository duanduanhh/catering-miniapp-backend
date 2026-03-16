package router

import (
	"github.com/gin-gonic/gin"
)

func InitCommonRouter(deps RouterDeps, r *gin.RouterGroup) {
	// 通用接口，无需认证
	commonRouter := r.Group("/")
	{
		commonRouter.GET("/close_reasons", deps.JobHandler.GetCloseReasons)
	}
}
