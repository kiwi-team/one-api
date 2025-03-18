package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func GetVersion(c *gin.Context) {
	version := "v1.0.2"
	fmt.Printf("version %v ,time %v\n", version, time.Now())
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    version,
	})
}
