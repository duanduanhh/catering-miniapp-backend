package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitEnterpriseRouter(deps RouterDeps, r *gin.RouterGroup) {
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/enterprise/ocr", deps.EnterpriseHandler.OCR)
		strictAuthRouter.POST("/enterprise/create", deps.EnterpriseHandler.Create)
		strictAuthRouter.POST("/enterprise/update", deps.EnterpriseHandler.Update)
		strictAuthRouter.DELETE("/enterprise/:id", deps.EnterpriseHandler.Delete)
		strictAuthRouter.POST("/enterprise/set_default", deps.EnterpriseHandler.SetDefault)
		strictAuthRouter.GET("/enterprise/my", deps.EnterpriseHandler.My)
		strictAuthRouter.GET("/enterprise/select_list", deps.EnterpriseHandler.SelectList)
	}

	noAuthRouter := r.Group("/")
	{
		noAuthRouter.GET("/enterprise/detail", deps.EnterpriseHandler.Detail)
	}
}
