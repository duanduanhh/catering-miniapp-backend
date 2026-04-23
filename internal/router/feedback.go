package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitFeedbackRouter(deps RouterDeps, r *gin.RouterGroup) {
	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger))
	{
		strictAuthRouter.POST("/feedback/submit", deps.FeedbackHandler.Submit)
		strictAuthRouter.POST("/feedback/my", deps.FeedbackHandler.List)
	}
}
