package monitor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/message"
	"github.com/songquanpeng/one-api/model"
)

func notifyRootUser(subject string, content string) {
	if config.MessagePusherAddress != "" {
		err := message.SendMessage(subject, content, content)
		if err != nil {
			logger.SysError(fmt.Sprintf("failed to send message: %s", err.Error()))
		} else {
			return
		}
	}
	if config.RootUserEmail == "" {
		config.RootUserEmail = model.GetRootUserEmail()
	}
	err := message.SendEmail(subject, config.RootUserEmail, content)
	if err != nil {
		logger.SysError(fmt.Sprintf("failed to send email: %s", err.Error()))
	}
}

// DisableChannel disable & notify
func DisableChannel(channelId int, channelName string, reason string) {
	model.UpdateChannelStatusById(channelId, model.ChannelStatusAutoDisabled)
	logger.SysLog(fmt.Sprintf("channel #%d has been disabled: %s", channelId, reason))
	subject := fmt.Sprintf("渠道「%s」（#%d）已被禁用", channelName, channelId)
	content := fmt.Sprintf("渠道「%s」（#%d）已被禁用，原因：%s", channelName, channelId, reason)
	notifyRootUser(subject, content)
}

func MetricDisableChannel(channelId int, successRate float64) {
	//model.UpdateChannelStatusById(channelId, model.ChannelStatusAutoDisabled)
	modelName := ""
	errlogs, _ := model.GetAllErrorLog(0, 10, channelId, modelName)
	channelName := ""
	errorMsg := "状态码		错误原因	时间  模型   <br>"
	if len(errlogs) > 0 {
		channelName = errlogs[0].ChannelName

		for _, log := range errlogs {
			errorMsg = errorMsg + fmt.Sprintf("%s	%s    %s	%s<br>", log.StatusCode, log.Message, time.Unix(log.CreatedAt, 0).Add(time.Hour*8).Format("2006-01-02 15:04:05"), log.ModelName)
		}
	}
	logger.SysLog(fmt.Sprintf("channel #%d has been disabled due to low success rate: %.2f", channelId, successRate*100))
	subject := fmt.Sprintf("渠道[%s] #%d  错误率太高，请留意", channelName, channelId)
	content := fmt.Sprintf("该渠道（#%d）在最近 %d 次调用中成功率为 %.2f%%，低于阈值 %.2f%%",
		channelId, config.MetricQueueSize, successRate*100, config.ChannelSuccessRateThreshold*100)
	content = content + "\n<br>\n" + errorMsg
	notifyRootUser(subject, content)
}

func MetricDisableChannelModel(key string, successRate float64) {
	//model.UpdateChannelStatusById(channelId, model.ChannelStatusAutoDisabled)
	arr := strings.Split(key, ";")
	channelIdStr := arr[0]
	channelId, err := strconv.Atoi(channelIdStr)
	if err != nil {
		logger.SysError(fmt.Sprintf("invalid channelId: %s", channelIdStr))
		return
	}
	modelId := arr[1] // 可能是擂台的模型id，也可能是模型的名称

	llmtestModel, err := model.GetLLMTestModel(modelId)
	if err != nil {
		logger.SysError(fmt.Sprintf("invalid modelId: %s,%v", modelId, err))
		return
	}
	modelName := llmtestModel.ModelName
	isArenaModel := false
	if modelName == "" {
		modelName = modelId
	} else {
		isArenaModel = true
	}
	model.UpdateAIModelChannelStatus(modelId, channelId, model.LLLMTEST_MODEL_CHANNEL_DISABLE)

	var errlogs []*model.ErrorLog
	total := 0

	if isArenaModel {
		total, _ = model.GetAIModelEnabledChannelCount(modelId, channelId)
		if total == -1 {
			errlogs, _ = model.GetAllErrorLog(0, 10, 0, modelName)
		} else {
			errlogs, _ = model.GetAllErrorLog(0, 10, channelId, modelName)
		}
	} else {
		errlogs, _ = model.GetAllErrorLog(0, 10, channelId, modelName)
	}
	channelName := ""
	errorMsg := "状态码		错误原因	时间    模型 <br>"
	subject := ""
	content := ""
	if len(errlogs) > 0 {
		channelName = errlogs[0].ChannelName

		for _, log := range errlogs {
			errorMsg = errorMsg + fmt.Sprintf("%s	%s    %s	%s<br>", log.StatusCode, log.Message, time.Unix(log.CreatedAt, 0).Add(time.Hour*8).Format("2006-01-02 15:04:05"), log.ModelName)
		}
	}

	logger.SysLog(fmt.Sprintf("channel #%d has been disabled due to low success rate: %.2f", channelId, successRate*100))
	if isArenaModel && total == -1 {
		// 模型关联的所有渠道，都被禁用了。这个时候，只能禁用擂台模型了
		model.UpdateLLMTestModel(modelId, model.LLLMTEST_MODEL_DISABLE)
		modelTypeName := model.GetArenaModelTypeName(llmtestModel.ModelType)
		subject = fmt.Sprintf("擂台模型【%s-%s】绑定的所有渠道错误率都太高, 已被禁用", modelName, modelTypeName)
		content = content + "\n<br>\n" + errorMsg
		msg, _ := model.SaveDisableOperateLog(0, 0, "", modelName, modelId, content, subject)
		subject = subject + msg
	} else {
		content = fmt.Sprintf("该渠道（#%d） 该模型(%s)在最近 %d 次调用中成功率为 %.2f%%，低于阈值 %.2f%%",
			channelId, modelName, config.MetricQueueSize, successRate*100, config.MetricSuccessRateThreshold*100)
		content = content + "\n<br>\n" + errorMsg
		subject = fmt.Sprintf("渠道[%s] #%d 模型[%s] 错误率太高，请留意", channelName, channelId, modelName)
	}
	notifyRootUser(subject, content)
}

// EnableChannel enable & notify
func EnableChannel(channelId int, channelName string) {
	model.UpdateChannelStatusById(channelId, model.ChannelStatusEnabled)
	logger.SysLog(fmt.Sprintf("channel #%d has been enabled", channelId))
	subject := fmt.Sprintf("渠道「%s」（#%d）已被启用", channelName, channelId)
	content := fmt.Sprintf("渠道「%s」（#%d）已被启用", channelName, channelId)
	notifyRootUser(subject, content)
}
