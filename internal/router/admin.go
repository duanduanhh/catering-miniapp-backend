package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitAdminRouter(deps RouterDeps, r *gin.RouterGroup) {
	adminGroup := r.Group("/admin")

	// 公开：登录
	adminGroup.POST("/login", deps.AdminAuthHandler.Login)

	// 受保护：所有其它 /admin/* 接口必须带 token header
	authed := adminGroup.Group("")
	authed.Use(middleware.AdminAuth(deps.Config, deps.Logger))
	{
		jobs := authed.Group("/jobs")
		jobs.POST("/list", deps.AdminJobHandler.List)
		jobs.POST("/disable", deps.AdminJobHandler.Disable)
		jobs.POST("/enable", deps.AdminJobHandler.Enable)
		jobs.POST("/delete", deps.AdminJobHandler.Delete)

		authed.POST("/users/list", deps.AdminListHandler.ListUsers)
		authed.POST("/enterprises/list", deps.AdminListHandler.ListEnterprises)
		authed.POST("/feedbacks/list", deps.AdminListHandler.ListFeedbacks)
		authed.POST("/contact_histories/list", deps.AdminListHandler.ListContactHistories)
		authed.POST("/reports/list", deps.AdminListHandler.ListReports)
	}
}
