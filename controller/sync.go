package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/model"
)

// 手动触发 同步数据到缓存
func SyncDataToCache(c *gin.Context) {
	model.SyncOptionsNow()
	model.SyncChannelCacheNow()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
