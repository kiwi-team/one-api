package monitor

import (
	"fmt"
	"time"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
)

/*
1. 10次成功率低于50%，禁用10分钟，然后10分钟后放开（不用测试）
2. 一天禁用 2次10钟后，第3次则直接触发6小时
3. 如果一天触发了1次6小时，第2次触发6小时的时候，就直接禁用，等待手动恢复
*/

// 因为成功率太低，禁用擂台模型
// modeId 可能是擂台模型Id，也可能是其他业务模型的名称
func MetricDisableModel(modelId string, successRate float64) {
	//model.UpdateChannelStatusById(channelId, model.ChannelStatusAutoDisabled)
	llmtestModel, err := model.GetLLMTestModel(modelId)
	if err != nil {
		logger.SysError(fmt.Sprintf("invalid modelId: %s,%v", modelId, err))
		return
	}
	modelName := llmtestModel.ModelName
	disableArenaModel := false
	if modelName == "" {
		modelName = modelId
	} else {
		disableArenaModel = true
	}
	errlogs, _ := model.GetAllErrorLog(0, 10, 0, modelName)
	errorMsg := "状态码		错误原因	时间     <br>"
	if len(errlogs) > 0 {
		for _, log := range errlogs {
			errorMsg = errorMsg + fmt.Sprintf("%s	%s    %s<br>", log.StatusCode, log.Message, time.Unix(log.CreatedAt, 0).Add(time.Hour*8).Format("2006-01-02 15:04:05"))
		}
	}
	logger.SysLog(fmt.Sprintf("model #%s has been disabled due to low success rate: %.2f", modelId, successRate*100))
	subject := fmt.Sprintf("模型[%s] 错误率太高，请留意", modelName)
	content := fmt.Sprintf("该模型 在最近 %d 次调用中成功率为 %.2f%%，低于阈值 %.2f%%",
		config.MetricQueueSize, successRate*100, config.ModelSuccessRateThreshold*100)
	content = content + "\n<br>\n" + errorMsg
	if disableArenaModel {
		model.UpdateLLMTestModel(modelId, model.LLLMTEST_MODEL_DISABLE)
		subject = fmt.Sprintf("擂台模型[%s] 错误率太高 ", llmtestModel.CompanyName+"/"+llmtestModel.ModelName+"/"+llmtestModel.Nickname)
		msg, _ := model.SaveDisableOperateLog(0, 0, "", modelName, modelId, content, subject)
		subject = subject + msg
	}
	notifyRootUser(subject, content)
}

func AutoEnable() {
	// 自动更新
	for {
		time.Sleep(1 * time.Minute)
		model.AutoEnableModelChannel()
		subject, content := model.AutoEnableArenaModel()
		if len(subject) > 0 && len(content) > 0 {
			notifyRootUser(subject, content)
		}
	}
}
