package imaginepro

import (
	"encoding/base64"
	"encoding/json"
	"errors"
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
	responseFormat := c.GetString("response_format")

	var imageResponse ImageResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &imageResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	if !imageResponse.Success {
		logger.SysError("imagineProAsyncTask err: " + string(responseBody))
		return openai.ErrorWrapper(errors.New(imageResponse.Error), "imagine_pro_async_task_failed", http.StatusInternalServerError), nil
	}

	taskResponse, _, err := asyncTaskWait(imageResponse.MessageID, apiKey)

	if err != nil {
		return openai.ErrorWrapper(err, "imagine_pro_async_task_wait_failed", http.StatusInternalServerError), nil
	}

	if taskResponse.Status != "DONE" {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: taskResponse.Error,
				Type:    "imagine_pro_error",
				Param:   "",
				Code:    taskResponse.Progress,
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	fullTextResponse := responseAli2OpenAIImage(taskResponse, responseFormat)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)
	return nil, nil
}

func asyncTask(messageId string, key string) (*TaskResponse, error, []byte) {
	url := fmt.Sprintf("https://api.imaginepro.ai/api/v1/midjourney/message/%s", messageId)

	var taskResponse TaskResponse

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return &taskResponse, err, nil
	}

	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.SysError("imagineProAsyncTask client.Do err: " + err.Error())
		return &taskResponse, err, nil
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body)

	var response TaskResponse
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		logger.SysError("aliAsyncTask NewDecoder err: " + err.Error())
		return &taskResponse, err, nil
	}

	return &response, nil, responseBody
}

func asyncTaskWait(messageId string, key string) (*TaskResponse, []byte, error) {
	waitSeconds := 2
	step := 0
	maxStep := 20

	var taskResponse TaskResponse
	var responseBody []byte

	for {
		step++
		rsp, err, body := asyncTask(messageId, key)
		responseBody = body
		if err != nil {
			return &taskResponse, responseBody, err
		}

		if rsp.Status == "" {
			return &taskResponse, responseBody, nil
		}
		//PROCESSING, QUEUED, DONE, FAIL

		switch rsp.Status {
		case "PROCESSING", "QUEUED":
			fmt.Printf("%s\n", rsp.Status)
		case "SUCCEEDED", "FAIL":
			return rsp, responseBody, nil
		}
		if step >= maxStep {
			break
		}
		time.Sleep(time.Duration(waitSeconds) * time.Second)
	}

	return nil, nil, fmt.Errorf("imagineProAsyncTaskWait timeout")
}

func responseAli2OpenAIImage(response *TaskResponse, responseFormat string) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}

	var b64Json string
	if responseFormat == "b64_json" {
		// 读取 data.Url 的图片数据并转存到 b64Json
		imageData, err := getImageData(response.Uri)
		if err != nil {
			// 处理获取图片数据失败的情况
			logger.SysError("getImageData Error getting image data: " + err.Error())
		}

		// 将图片数据转为 Base64 编码的字符串
		b64Json = Base64Encode(imageData)
	} else {
		// 如果 responseFormat 不是 "b64_json"，则直接使用 data.B64Image
		b64Json = ""
	}

	imageResponse.Data = append(imageResponse.Data, openai.ImageData{
		Url:           response.Uri,
		B64Json:       b64Json,
		RevisedPrompt: "",
	})
	return &imageResponse
}

func getImageData(url string) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	imageData, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return imageData, nil
}

func Base64Encode(data []byte) string {
	b64Json := base64.StdEncoding.EncodeToString(data)
	return b64Json
}
