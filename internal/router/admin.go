package router

import "github.com/gin-gonic/gin"

func InitAdminRouter(deps RouterDeps, r *gin.RouterGroup) {
	adminGroup := r.Group("/admin")
	{
		jobs := adminGroup.Group("/jobs")
		jobs.POST("/list", deps.AdminJobHandler.List)
		jobs.POST("/disable", deps.AdminJobHandler.Disable)
		jobs.POST("/enable", deps.AdminJobHandler.Enable)
		jobs.POST("/delete", deps.AdminJobHandler.Delete)
	}
}
