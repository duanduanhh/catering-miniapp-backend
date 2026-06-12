package router

import (
	"github.com/gin-gonic/gin"

	"github.com/go-nunu/nunu-layout-advanced/internal/middleware"
)

func InitContactFeedbackRouter(deps RouterDeps, r *gin.RouterGroup) {
	noAuthRouter := r.Group("/")
	{
		noAuthRouter.GET("/contact_feedback/reasons", deps.ContactFeedbackHandler.Reasons)
	}

	strictAuthRouter := r.Group("/").Use(middleware.StrictAuth(deps.JWT, deps.Logger, deps.UserRepo))
	{
		strictAuthRouter.POST("/contact_feedback/submit", deps.ContactFeedbackHandler.Submit)
	}
}
