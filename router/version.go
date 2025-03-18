package router

import (
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/controller"
)

func SetVersionRouter(router *gin.Engine) {
	versionRouter := router.Group("/version")
	versionRouter.Use()
	{
		versionRouter.GET("", controller.GetVersion)
	}
}
