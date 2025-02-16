package model

import (
	"errors"
	"fmt"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"gorm.io/gorm"
)

// 系统禁用启用操作日志
type OperateLog struct {
	Id             int    `json:"id"`
	UserId         int    `json:"user_id" gorm:"index"`
	CreatedDay     string `json:"created_day" gorm:"default:''"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint;index:idx_created_a"`
	ChannelId      int    `json:"channel_id" gorm:"index"`
	ChannelName    string `json:"channel_name" gorm:"default:''"`
	ModelName      string `json:"model_name" gorm:"default:''"`
	ModelId        string `json:"model_id" gorm:"default:''"`
	Action         string `json:"action" gorm:"default:''"`          // disable enable
	DisableTime    string `json:"disable_time" gorm:"default:''"`    // 禁用时间 10m 6h这样的字符串
	EnableTime     int64  `json:"enable_time" gorm:"default:0"`      // 启用时间
	NextEnableTime int64  `json:"next_enable_time" gorm:"default:0"` // 禁用的时候，记录下次启用时间
	EmailContent   string `json:"email_content" gorm:"default:''"`   // 需要发送的邮件内容
	EmailSubject   string `json:"email_subject" gorm:"default:''"`   // 邮件的标题
}

type DisableCountInfo struct {
	TenMinutes int `json:"ten_minutes" default:"0"`
	SixHours   int `json:"six_hours" default:"0"`
}

func SaveDisableOperateLog(userId int, channelId int, channelName string, modelName string, modelId string, emailContent string, emailSubject string) (string, error) {
	action := "disable"
	day := helper.GetDayString()
	var len int64
	err := DB.Model(&OperateLog{}).Where("model_id = ? and created_day = ?", modelId, day).Count(&len).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.SysError("failed to get last operate_log: " + err.Error())
		return "", err
	}
	disableTime := ""
	var nextEnableTime int64
	msg := fmt.Sprintf("一天之内，被禁用【%d】次 ", len+1)
	switch len {
	case 0, 1:
		disableTime = "10m"
		nextEnableTime = helper.GetTimestamp() + 10*60
		msg += " 10分钟后自动启用"
	case 2:
		disableTime = "6h"
		nextEnableTime = helper.GetTimestamp() + 6*60*60
		msg += " 6小时后自动启用"
	case 3:
		fallthrough
	default:
		disableTime = "60d"
		nextEnableTime = helper.GetTimestamp() + 600*24*60*60
		msg += " 需要手动启用！"
	}
	log := &OperateLog{
		UserId:         userId,
		CreatedAt:      helper.GetTimestamp(),
		CreatedDay:     day,
		ChannelId:      channelId,
		ChannelName:    channelName,
		ModelName:      modelName,
		ModelId:        modelId,
		Action:         action,
		DisableTime:    disableTime,
		NextEnableTime: nextEnableTime,
		EmailContent:   emailContent,
		EmailSubject:   msg + emailSubject,
	}
	err1 := DB.Create(log).Error
	if err1 != nil {
		logger.SysError("failed to record operate_log: " + err1.Error())
	}
	return msg, err1
}

// 系统启用操作日志
func SaveEnableOperateLog(id int) error {
	err := DB.Model(&OperateLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"enable_time": helper.GetTimestamp(),
		"action":      "enable",
	}).Error
	if err != nil {
		logger.SysError("failed to record operate_log: " + err.Error())
	}
	return err
}
