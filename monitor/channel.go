package monitor

import (
	"fmt"
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
	errlogs, _ := model.GetAllErrorLog(0, 10, channelId)
	channelName := ""
	errorMsg := "错误原因	时间     状态码<br>"
	if len(errlogs) > 0 {
		channelName = errlogs[0].ChannelName

		for _, log := range errlogs {
			errorMsg = errorMsg + fmt.Sprintf("%s	%s    %s<br>", log.StatusCode, log.Message, time.Unix(log.CreatedAt, 0).Format("2006-01-02 15:04:05"))
		}
	}
	logger.SysLog(fmt.Sprintf("channel #%d has been disabled due to low success rate: %.2f", channelId, successRate*100))
	subject := fmt.Sprintf("渠道[%s] #%d 错误率太高，请留意", channelName, channelId)
	content := fmt.Sprintf("该渠道（#%d）在最近 %d 次调用中成功率为 %.2f%%，低于阈值 %.2f%%",
		channelId, config.MetricQueueSize, successRate*100, config.MetricSuccessRateThreshold*100)
	content = content + "\n<br>\n" + errorMsg
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
