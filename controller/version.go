package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
)

func GetVersion(c *gin.Context) {
	version := "v1.0.3"
	logger.Info(c, fmt.Sprintf("version %v ,time %v\n", version, time.Now()))
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    version,
	})
}
