package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitReportRouter(deps RouterDeps, r *gin.RouterGroup) {
	noAuthRouter := r.Group("/")
	{
		noAuthRouter.GET("/report/reasons", deps.ReportHandler.Reasons)
	}

	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger, deps.UserRepo))
	{
		strictAuthRouter.POST("/report/submit", deps.ReportHandler.Submit)
	}
}
