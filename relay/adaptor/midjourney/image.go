package midjourney

// https://omniai.pandalla.ai/guide/#midjourney-api-%E4%BD%BF%E7%94%A8%E6%8C%87%E5%8D%97

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	var mjTaskResponse MJImageSubmitTaskResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	err = json.Unmarshal(responseBody, &mjTaskResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	mjResponse, err := asyncTaskWait(mjTaskResponse.Result, apiKey)

	if err != nil {
		return openai.ErrorWrapper(err, "klingai_async_task_wait_failed", http.StatusInternalServerError), nil
	}

	if mjResponse.Status != "SUCCESS" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: mjResponse.FailReason,
				Type:    "panda_midjourney_error",
				Param:   "",
				Code:    mjResponse.ID,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := response2OpenAI(mjResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return nil, nil
}

func asyncTaskQuery(taskId, token string) (*MJImageQueryTaskResponse, error) {
	// 获取任务状态
	url := fmt.Sprintf("https://api.pandalla.ai/mj-fast/mj/task/%s/fetch", taskId)

	var taskResponse MJImageQueryTaskResponse

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &taskResponse, err
	}

	req.Header.Set("mj-api-secret", token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.SysError("panda Midjourney AsyncTask client.Do err: " + err.Error())
		return &taskResponse, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.SysError("panda MidjourneyAsyncTaskQuery io.ReadAll err: " + err.Error())
		return &taskResponse, err
	}

	err = json.Unmarshal(responseBody, &taskResponse)
	if err != nil {
		logger.SysError("panda Midjourney AsyncTask NewDecoder err: " + err.Error())
		return &taskResponse, err
	}

	logger.SysLog("panda Midjourney AsyncTask response: " + string(responseBody))

	return &taskResponse, nil
}

func asyncTaskWait(taskId, apiKey string) (*MJImageQueryTaskResponse, error) {
	waitSeconds := 3
	step := 0
	maxStep := 40

	var taskResponse MJImageQueryTaskResponse

	for {
		step++
		rsp, err := asyncTaskQuery(taskId, apiKey)
		if err != nil {
			return &taskResponse, err
		}

		switch rsp.Status {
		case "FAILURE":
			fallthrough
		case "SUCCESS":
			return rsp, nil
		}

		if step >= maxStep {
			break
		}

		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	return nil, fmt.Errorf("pandaMJAsyncTaskWait timeout")
}

func response2OpenAI(response *MJImageQueryTaskResponse) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}

	imageResponse.Data = append(imageResponse.Data, openai.ImageData{
		Url:           response.ImageURL,
		B64Json:       "",
		RevisedPrompt: response.PromptEn,
	})
	return &imageResponse
}
