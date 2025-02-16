package ailab

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
	"github.com/songquanpeng/one-api/relay/adaptor/openai" // Ensure this path is correct and the package exists
	"github.com/songquanpeng/one-api/relay/model"
)

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	apiKey := c.Request.Header.Get("Authorization")
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")

	var mjTaskResponse ImageSubmitTaskResponse
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
	if mjTaskResponse.Code != 0 {
		return openai.ErrorWrapper(fmt.Errorf(mjTaskResponse.Msg, mjTaskResponse.Code), "ailab_image_error", http.StatusInternalServerError), nil
	}

	imgResponse, err := asyncTaskWait(mjTaskResponse.Data.TaskID, apiKey)

	if err != nil {
		return openai.ErrorWrapper(err, "ailab_async_task_wait_failed", http.StatusInternalServerError), nil
	}

	if imgResponse.Data.State != 2 {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: imgResponse.Msg,
				Type:    "ailab_image_error",
				Param:   "",
				Code:    imgResponse.Code,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := response2OpenAI(imgResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return nil, nil
}

func asyncTaskQuery(taskId int, token string) (*ImageQueryTaskResponse, error) {
	// 获取任务状态
	url := fmt.Sprintf("https://magic-animation.intern-ai.org.cn/magic-animation-be/api/v1/infer_tasks/%d", taskId)

	var taskResponse ImageQueryTaskResponse

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &taskResponse, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.SysError("ailab image AsyncTask client.Do err: " + err.Error())
		return &taskResponse, err
	}

	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		logger.SysError("ailab image AsyncTaskQuery io.ReadAll err: " + err.Error())
		return &taskResponse, err
	}

	err = json.Unmarshal(responseBody, &taskResponse)
	if err != nil {
		logger.SysError("ailab image  AsyncTask NewDecoder err: " + err.Error())
		return &taskResponse, err
	}

	logger.SysLog("ailab image  AsyncTask response: " + string(responseBody))

	return &taskResponse, nil
}

func asyncTaskWait(taskId int, apiKey string) (*ImageQueryTaskResponse, error) {
	waitSeconds := 3
	step := 0
	maxStep := 40

	var taskResponse ImageQueryTaskResponse

	for {
		step++
		rsp, err := asyncTaskQuery(taskId, apiKey)
		if err != nil {
			return &taskResponse, err
		}

		switch rsp.Data.State {
		case 0, 1:
			fmt.Println("ailab image  state: ", rsp.Data.State)
		case 2:
			if rsp.Data.Output != "" {
				return rsp, nil
			} else {
				return nil, fmt.Errorf("ailab image  output is empty")
			}
		}

		if step >= maxStep {
			break
		}

		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	return nil, fmt.Errorf("ailab image  timeout")
}

func response2OpenAI(response *ImageQueryTaskResponse) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}

	imageResponse.Data = append(imageResponse.Data, openai.ImageData{
		Url:           response.Data.Output,
		B64Json:       "",
		RevisedPrompt: "",
	})
	return &imageResponse
}
