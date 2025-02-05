package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
)

const (
	LLLMTEST_MODEL_ENABLED         = 1
	LLLMTEST_MODEL_DISABLE         = 2
	LLLMTEST_MODEL_CHANNEL_ENABLED = 1
	LLLMTEST_MODEL_CHANNEL_DISABLE = 2
)

type LLMTestModel struct {
	Id          int    `json:"id"`
	ModelId     string `json:"modelId" gorm:"column:modelId"`
	ModelType   int    `json:"modelType" gorm:"column:modelType"`
	CompanyName string `json:"companyName" gorm:"column:companyName"`
	ModelName   string `json:"modelName" gorm:"column:modelName"`
	Nickname    string `json:"nickname" gorm:"column:nickname"`
	OneApiKey   string `json:"oneApiKey" gorm:"column:oneApiKey"`
	Status      int    `json:"status" gorm:"column:status"`
}

type TotalRes struct {
	Total int `json:"total" gorm:"column:total"`
}

type OpItem struct {
	Id             int    `json:"id"`
	ModeId         string `json:"modelId" gorm:"column:modelId"`
	NextEnableTime int64  `json:"nextEnableTime" gorm:"column:nextEnableTime"`
}

func GetArenaModelTypeName(modelType int) string {
	name := ""
	switch modelType {
	case 1:
		name = "文生文"
	case 2:
		name = "文生图"
	case 3:
		name = "图文擂台"
	case 4:
		name = "prompt PK"
	case 5:
		name = "场景PK"
	default:
		// 未知模型
	}
	return name

}

func GetLLMTestModel(modelId string) (*LLMTestModel, error) {
	var model *LLMTestModel
	sql := `select id,"modelId","modelType","companyName","modelName","nickname","status","oneApiKey" from "AIModel" where "modelId" = ?`
	err := LLMTEST_DB.Raw(sql, modelId).Scan(&model).Error
	return model, err
}

func UpdateLLMTestModel(modelId string, status int) error {
	sql := `update "AIModel" set status = ?,"changeStatusAt"= ? where "modelId" = ?`
	changeStatusAt := helper.FormatTime(time.Now()) //10分钟后，这个渠道再次启用
	err := LLMTEST_DB.Exec(sql, status, changeStatusAt, modelId).Error
	return err
}

func UpdateAIModelChannelStatus(modelId string, channelId int, status int) error {
	if status == LLLMTEST_MODEL_CHANNEL_ENABLED {
		sql := `update "AIModelChannels" set status = ?  where "modelId" = ? and "channelId" = ? and deleted = 0`
		err := LLMTEST_DB.Exec(sql, status, modelId, channelId).Error
		return err
	} else {
		nextEnableTime := helper.FormatTime(time.Now().Add(time.Minute * 10)) //10分钟后，这个渠道再次启用
		sql := `update "AIModelChannels" set status = ?,"nextEnableTime" = ? where "modelId" = ? and "channelId" = ? and deleted = 0`
		err := LLMTEST_DB.Exec(sql, status, nextEnableTime, modelId, channelId).Error
		return err
	}
}

func GetAIModelEnabledChannelCount(modelId string, channelId int) (int, error) {
	var count *TotalRes
	sql := `select count(1) as total from "" where "modelId" = ? and "channelId" = ?`
	err := LLMTEST_DB.Raw(sql, modelId, channelId).Scan(&count).Error
	if err != nil {
		return 0, err
	}
	if count.Total < 1 {
		return 0, nil
	} else {
		sql := `select count(1) as total from "AIModelChannels" where "modelId" = ? and "channelId" = ? and status = ? and deleted = 0`
		err := LLMTEST_DB.Raw(sql, modelId, channelId, LLLMTEST_MODEL_CHANNEL_ENABLED).Scan(&count).Error
		if err != nil {
			return 0, err
		}
		if count.Total < 1 {
			// 擂台模型关联了多个渠道，但是没有一个渠道可用了。
			return -1, errors.New("all channel is disabled")
		}
		return count.Total, nil
	}
}

func AutoEnableModelChannel() {
	sql := `update "AIModelChannels" set status = 1 where "nextEnableTime" < ? and status = 2 and deleted = 0`
	err := LLMTEST_DB.Exec(sql, helper.FormatTime(time.Now())).Error
	if err != nil {
		logger.SysError("failed to auto enable AIModelChannels: " + err.Error())
	}
}

func AutoEnableArenaModel() (string, string) {
	action := "disable"
	day := helper.GetDayString()
	yesterDay := helper.GetYesterDayString()
	var opList []*OpItem
	sql := `select max(id) as id, "model_id" as "modelId",max("next_enable_time") as "nextEnableTime"  from "operate_logs" where action = ? and ("created_day" = ? or "created_day" = ?) group by "model_id"`
	err := DB.Raw(sql, action, day, yesterDay).Scan(&opList).Error
	if err != nil {
		logger.SysError("failed to auto enable AutoEnableArenaModel: " + err.Error())
	}
	now := helper.GetTimestamp()
	subject := ""
	content := ""
	i := 0
	for _, op := range opList {
		if op.NextEnableTime < now {
			err = UpdateLLMTestModel(op.ModeId, LLLMTEST_MODEL_ENABLED)
			if err != nil {
				logger.SysError("failed to UpdateLLMTestModel: " + err.Error())
			}
			SaveEnableOperateLog(op.Id)
			arenaModel, err := GetLLMTestModel(op.ModeId)
			if err != nil {
				logger.SysError("failed to GetLLMTestModel: " + op.ModeId + err.Error())
				continue
			}
			i = i + 1
			subject = fmt.Sprintf("已自动启用 %d 个擂台模型", i)
			modelTypeName := GetArenaModelTypeName(arenaModel.ModelType)
			content = content + fmt.Sprintf("渠道模型【%s-%s】已自动启用\n", modelTypeName, arenaModel.CompanyName+"/"+arenaModel.ModelName+"/"+arenaModel.Nickname)
		}
	}
	return subject, content
}
