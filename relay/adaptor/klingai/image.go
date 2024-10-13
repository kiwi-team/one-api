package klingai

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
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = apiKey[len("Bearer "):]

	var klingaiTaskResponse TaskResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}

	err = json.Unmarshal(responseBody, &klingaiTaskResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	klingaiResponse, _, err := asyncTaskWait(klingaiTaskResponse.Data.TaskId, apiKey)

	if err != nil {
		return openai.ErrorWrapper(err, "klingai_async_task_wait_failed", http.StatusInternalServerError), nil
	}

	if klingaiResponse.Data.TaskStatus != "succeed" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: klingaiResponse.Message,
				Type:    "klingai_error",
				Param:   "",
				Code:    klingaiResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := responseKlingai2OpenAI(klingaiResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return nil, nil
}

func asyncTaskQuery(taskId, token string) (*TaskResponse, []byte, error) {
	url := fmt.Sprintf("https://api.klingai.com/v1/images/generations/%s", taskId)

	var taskResponse TaskResponse

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &taskResponse, nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.SysError("klingai AsyncTask client.Do err: " + err.Error())
		return &taskResponse, nil, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.SysError("klingai AsyncTask io.ReadAll err: " + err.Error())
		return &taskResponse, nil, err
	}

	var response TaskResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		logger.SysError("klingai AsyncTask NewDecoder err: " + err.Error())
		return &taskResponse, nil, err
	}

	logger.SysLog("klingai AsyncTask response: " + string(responseBody))

	return &response, responseBody, nil
}

func asyncTaskWait(taskId, apiKey string) (*TaskResponse, []byte, error) {
	waitSeconds := 2
	step := 0
	maxStep := 20

	var taskResponse TaskResponse
	var responseBody []byte

	token, err := GetToken(apiKey)
	if err != nil {
		return nil, nil, err
	}

	for {
		step++
		rsp, body, err := asyncTaskQuery(taskId, token)
		responseBody = body
		if err != nil {
			return &taskResponse, responseBody, err
		}

		if rsp.Data.TaskStatus == "" {
			return &taskResponse, responseBody, nil
		}

		switch rsp.Data.TaskStatus {
		case "failed":
			fallthrough
		case "succeed":
			return rsp, responseBody, nil
		}

		if step >= maxStep {
			break
		}

		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	return nil, nil, fmt.Errorf("klingaiAsyncTaskWait timeout")
}

func responseKlingai2OpenAI(response *TaskResponse) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}

	logger.SysLog("klingai response images count: " + fmt.Sprint(len(response.Data.TaskResult.Images)))

	for _, data := range response.Data.TaskResult.Images {
		logger.SysLog("klingai response data url: " + data.Url)
		imageResponse.Data = append(imageResponse.Data, openai.ImageData{
			Url:           data.Url,
			B64Json:       "",
			RevisedPrompt: "",
		})
	}

	return &imageResponse
}
