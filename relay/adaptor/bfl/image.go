package bfl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	apiKey := c.Request.Header.Get("X-Key")

	var bflTaskResponse TaskResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}

	err = resp.Body.Close()

	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	err = json.Unmarshal(responseBody, &bflTaskResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	bflResponse, _, err := asyncTaskWait(bflTaskResponse.Id, apiKey)

	if err != nil {
		return openai.ErrorWrapper(err, "bfl_async_task_wait_failed", http.StatusInternalServerError), nil
	}

	if bflResponse.Status != "Ready" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: bflResponse.Status,
				Type:    "bfl_error",
				Param:   "",
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := responseBfl2OpenAI(bflResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return nil, nil
}

func asyncTaskQuery(taskId, apiKey string) (*TaskResponse, []byte, error) {
	url := fmt.Sprintf("https://api.bfl.ml/v1/get_result?id=%s", taskId)

	var taskResponse TaskResponse

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &taskResponse, nil, err
	}

	req.Header.Set("X-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return &taskResponse, nil, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.SysError("bfl asyncTaskQuery io.ReadAll: " + err.Error())
		return &taskResponse, nil, err
	}

	var response TaskResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		logger.SysError("bfl asyncTaskQuery json.Unmarshal: " + err.Error())
		return &taskResponse, nil, err
	}

	logger.SysLog("bfl asyncTaskQuery response: " + string(responseBody))

	return &response, responseBody, nil
}

func asyncTaskWait(taskId, apiKey string) (*TaskResponse, []byte, error) {
	waitSeconds := 2
	step := 0
	maxStep := 20

	var taskResponse TaskResponse
	var responseBody []byte

	for {
		step++
		rsp, body, err := asyncTaskQuery(taskId, apiKey)
		responseBody = body
		if err != nil {
			return &taskResponse, responseBody, err
		}

		if rsp.Status == "" {
			return &taskResponse, responseBody, nil
		}

		switch rsp.Status {
		case "Task not found":
			fallthrough
		case "Content Moderated":
			fallthrough
		case "Request Moderated":
			fallthrough
		case "Error":
			fallthrough
		case "Ready":
			return rsp, responseBody, nil
		}

		if step >= maxStep {
			break
		}

		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	return nil, nil, fmt.Errorf("bfl asyncTaskWait timeout")
}

func responseBfl2OpenAI(bflResponse *TaskResponse) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}

	imageResponse.Data = append(imageResponse.Data, openai.ImageData{
		Url:           bflResponse.Result.Sample,
		B64Json:       "",
		RevisedPrompt: "",
	})

	return &imageResponse
}
