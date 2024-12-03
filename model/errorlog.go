package model

import (
	"fmt"
	"strconv"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/model"
)

type ErrorLog struct {
	Id          int    `json:"id"`
	UserId      int    `json:"user_id" gorm:"index"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;index:idx_created_a"`
	ChannelId   int    `json:"channel" gorm:"index"`
	ChannelName string `json:"channel_name" gorm:"default:''"`
	Message     string `json:"message" gorm:"default:''"`
	Type        string `json:"type" gorm:"default:''"`
	Param       string `json:"param" gorm:"default:''"`
	Code        string `json:"code" gorm:"default:''"`
	StatusCode  string `json:"status_code" gorm:"default:''"`
	Body        string `json:"body" gorm:"default:''"`
}

func GetAllErrorLog(startIdx int, num int, channelId int) ([]*ErrorLog, error) {
	var errorLogs []*ErrorLog
	var err error
	if channelId > 0 {
		err = DB.Order("id desc").Limit(num).Offset(startIdx).Find(&errorLogs).Error
	} else {
		err = DB.Order("id desc").Where("channel = ?", channelId).Limit(num).Offset(startIdx).Find(&errorLogs).Error
	}
	return errorLogs, err
}

func SaveErrorLog(userId int, channelId int, channelName string, err *model.ErrorWithStatusCode, body string) error {
	log := &ErrorLog{
		UserId:      userId,
		CreatedAt:   helper.GetTimestamp(),
		ChannelId:   channelId,
		Message:     err.Message,
		Type:        err.Type,
		Param:       err.Param,
		ChannelName: channelName,
		Code:        fmt.Sprintf("%v", err.Code),
		StatusCode:  strconv.Itoa(err.StatusCode),
		Body:        body,
	}
	err1 := DB.Create(log).Error
	if err1 != nil {
		logger.SysError("failed to record error_log: " + err1.Error())
	}
	return err1
}
