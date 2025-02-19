package monitor

import (
	"fmt"
	"os"

	"github.com/songquanpeng/one-api/common/config"
)

/*
需要按照渠道+模型的维度来统计分析模型。如果最后模型关联的可用的渠道数量为零了,则直接禁用模型。
 - 渠道+模型(key)的维度 禁用逻辑：
	1. 近10次的成功率。如果定于50%，则对应模型关联的可用的渠道就减少一个。这个key被禁用10分钟
 - 擂台模型的禁用/启用逻辑：
	- 禁用
		1. 模型关联的可用的渠道数量为零了
		2. 擂台模型的10次成功率低于50%
			- 启用后，相关的渠道+模型都启用
		3. 一天禁用 1次10钟后,再启用
		4. 一天禁用 2次10钟后，第3次则直接触发6小时
		4. 如果一天触发了1次6小时，第2次触发6小时的时候，就直接禁用，等待手动恢复
	- 启用
*/

// 监控模型和渠道的成功率
var store = make(map[string][]bool)
var metricSuccessChan = make(chan string, config.MetricSuccessChanSize)
var metricFailChan = make(chan string, config.MetricFailChanSize)

// 监控模型的(llmtest模型的,擂台模型ID，或者其他业务的模型名称)
var modelStore = make(map[string][]bool)
var modelSuccessChan = make(chan string, config.MetricSuccessChanSize)
var modelFailChan = make(chan string, config.MetricFailChanSize)

// 监控渠道的
var channelStore = make(map[int][]bool)
var channelSuccessChan = make(chan int, config.MetricSuccessChanSize)
var channelFailChan = make(chan int, config.MetricFailChanSize)

func consumeSuccess(key string) {
	if len(store[key]) > config.MetricQueueSize {
		store[key] = store[key][1:]
	}
	store[key] = append(store[key], true)
}

func consumeModelSuccess(modelId string) {
	if len(modelStore[modelId]) > config.MetricQueueSize {
		modelStore[modelId] = modelStore[modelId][1:]
	}
	modelStore[modelId] = append(store[modelId], true)
}

func consumeChannelSuccess(channelId int) {
	if len(channelStore[channelId]) > config.MetricQueueSize {
		channelStore[channelId] = channelStore[channelId][1:]
	}
	channelStore[channelId] = append(channelStore[channelId], true)
}

func consumeChannelFail(key int) (bool, float64) {
	if len(channelStore[key]) > config.MetricQueueSize {
		channelStore[key] = channelStore[key][1:]
	}
	channelStore[key] = append(channelStore[key], false)
	successCount := 0
	for _, success := range channelStore[key] {
		if success {
			successCount++
		}
	}
	successRate := float64(successCount) / float64(len(channelStore[key]))
	if len(channelStore[key]) < config.MetricQueueSize {
		return false, successRate
	}
	if successRate < config.ChannelSuccessRateThreshold {
		channelStore[key] = make([]bool, 0)
		return true, successRate
	}
	return false, successRate
}

func consumeModelFail(key string) (bool, float64) {
	if len(modelStore[key]) > config.MetricQueueSize {
		modelStore[key] = modelStore[key][1:]
	}
	modelStore[key] = append(modelStore[key], false)
	successCount := 0
	for _, success := range modelStore[key] {
		if success {
			successCount++
		}
	}
	successRate := float64(successCount) / float64(len(modelStore[key]))
	if len(modelStore[key]) < config.MetricQueueSize {
		return false, successRate
	}
	if successRate < config.ModelSuccessRateThreshold {
		modelStore[key] = make([]bool, 0)
		return true, successRate
	}
	return false, successRate
}

func consumeFail(key string) (bool, float64) {
	if len(store[key]) > config.MetricQueueSize {
		store[key] = store[key][1:]
	}
	store[key] = append(store[key], false)
	successCount := 0
	for _, success := range store[key] {
		if success {
			successCount++
		}
	}
	successRate := float64(successCount) / float64(len(store[key]))
	if len(store[key]) < config.MetricQueueSize {
		return false, successRate
	}
	if successRate < config.MetricSuccessRateThreshold {
		store[key] = make([]bool, 0)
		return true, successRate
	}
	return false, successRate
}

func metricSuccessConsumer() {
	for {
		select {
		case channelId := <-metricSuccessChan:
			consumeSuccess(channelId)
		case modelId := <-modelSuccessChan:
			consumeModelSuccess(modelId)
		case channelId := <-channelSuccessChan:
			consumeChannelSuccess(channelId)
		}
	}
}

func metricFailConsumer() {
	for {
		select {
		case key := <-metricFailChan:
			disable, successRate := consumeFail(key)
			if disable && os.Getenv("LLMTEST_SQL_DSN") != "" {
				go MetricDisableChannelModel(key, successRate)
			}
		case modelId := <-modelFailChan:
			disable, successRate := consumeModelFail(modelId)
			if disable && os.Getenv("LLMTEST_SQL_DSN") != "" {
				go MetricDisableModel(modelId, successRate)
			}
		case channelId := <-channelFailChan:
			if disable, successRate := consumeChannelFail(channelId); disable {
				go MetricDisableChannel(channelId, successRate)
			}
		}
	}
}

func init() {
	if config.EnableMetric {
		go metricSuccessConsumer()
		go metricFailConsumer()
	}
}

/**
* modelId 可能是擂台的模型Id，如果是其他业务,modelId可能是空字符串
 */
func Emit(channelId int, modelName string, modelId string, success bool) {
	if !config.EnableMetric {
		return
	}
	key := fmt.Sprintf("%d;%s", channelId, modelId)
	if modelId == "" {
		key = fmt.Sprintf("%d;%s", channelId, modelName)
	}
	go func() {
		if success {
			metricSuccessChan <- key
			modelSuccessChan <- modelId // 可能是擂台的模型ID，也可能是其他模块的模型名称
			channelSuccessChan <- channelId
		} else {
			metricFailChan <- key
			modelFailChan <- modelId
			channelFailChan <- channelId
		}
	}()
}
