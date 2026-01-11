package router

import (
	"github.com/gin-gonic/gin"
	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitContactHistoryRouter(deps RouterDeps, r *gin.RouterGroup) {
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/contact_history/out", deps.ContactHistoryHandler.ListOut)
		strictAuthRouter.POST("/contact_history/in", deps.ContactHistoryHandler.ListIn)
	}
}
