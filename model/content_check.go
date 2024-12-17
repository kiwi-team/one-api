package model

import "github.com/songquanpeng/one-api/common/helper"

type ContentCheck struct {
	Id          int    `json:"id"`
	CreatedAt   int64  `json:"created_at" gorm:"bigint;index:idx_created_a"`
	ChannelId   int    `json:"channel" gorm:"index"`
	ChannelName string `json:"channel_name" gorm:"default:''"`
	Content     string `json:"content" gorm:"default:''"`
	Result      string `json:"result" gorm:"default:''"`
}

func InsertContentCheck(contentCheck *ContentCheck) error {
	data := DB.Create(&ContentCheck{
		ChannelId:   contentCheck.ChannelId,
		ChannelName: contentCheck.ChannelName,
		Content:     contentCheck.Content,
		Result:      contentCheck.Result,
		CreatedAt:   helper.GetTimestamp(),
	})
	return data.Error
}
