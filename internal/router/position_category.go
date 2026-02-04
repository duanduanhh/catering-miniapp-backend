package router

import (
	"github.com/gin-gonic/gin"
)

func InitPositionCategoryRouter(deps RouterDeps, r *gin.RouterGroup) {
	noAuthRouter := r.Group("/")
	{
		noAuthRouter.GET("/positions/all", deps.PositionCategoryHandler.GetAllCategories)
	}
}
