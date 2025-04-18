package sensenova

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/common/render"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func ConvertRequest(request model.GeneralOpenAIRequest) *ChatRequest {
	sensenovaMessages := make([]Message, len(request.Messages))
	for i, message := range request.Messages {
		contentArray := make([]ContentItem, 0)
		for _, content := range message.ParseContent() {
			if content.Type == model.ContentTypeText {
				contentArray = append(contentArray, ContentItem{
					Type: "text",
					Text: content.Text,
				})
			} else if content.Type == model.ContentTypeImageURL {
				if strings.HasPrefix(content.ImageURL.Url, "data:") {
					contentArray = append(contentArray, ContentItem{
						Type:        "image_base64",
						ImageBase64: content.ImageURL.Url,
					})
				} else {
					contentArray = append(contentArray, ContentItem{
						Type:     "image_url",
						ImageUrl: content.ImageURL.Url,
					})

				}
			} else if content.Type == model.ContentTypeVideoURL {
				contentArray = append(contentArray, ContentItem{
					Type:     "video_url",
					VideoUrl: content.VideoURL.Url,
				})
			}
		}
		sensenovaMessages[i] = Message{
			Role:    message.Role,
			Content: contentArray,
		}
	}
	chatRequest := &ChatRequest{
		Model:    request.Model,
		Messages: sensenovaMessages,
		MaxNewTokens: func() int {
			if request.MaxTokens > 0 {
				return request.MaxTokens
			}
			return 1024
		}(),
		Temperature: func() float64 {
			if request.Temperature != nil && *request.Temperature > 0 {
				return *request.Temperature
			}
			return 0.8
		}(),
		TopP: func() float64 {
			if request.TopP != nil && *request.TopP > 0 {
				return *request.TopP
			}
			return 0.95
		}(),
		RepetitionPenalty: 1.0,
		Stream:            request.Stream,
		User:              request.User,
	}
	return chatRequest
}

func Handler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	ctx := c.Request.Context()
	var newResponse ChatResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return openai.ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	logger.Debugf(ctx, "response body: %s\n", responseBody)
	err = json.Unmarshal(responseBody, &newResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	fullTextResponse := responseSensenova2OpenAI(&newResponse)
	jsonResponse, err := json.Marshal(fullTextResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, err = c.Writer.Write(jsonResponse)
	return nil, &fullTextResponse.Usage
}

func StreamHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var usage model.Usage
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		if i := strings.Index(string(data), "\n"); i >= 0 {
			return i + 1, data[0:i], nil
		}
		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	common.SetEventStreamHeaders(c)

	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < 5 || data[:5] != "data:" {
			continue
		}
		data = data[5:]

		var newResponse StreamResponse
		err := json.Unmarshal([]byte(data), &newResponse)
		if err != nil {
			logger.SysError("error unmarshalling stream response: " + err.Error())
			continue
		}
		if newResponse.Data.Usage.PromptTokens > usage.PromptTokens {
			usage.PromptTokens = newResponse.Data.Usage.PromptTokens
		}
		if newResponse.Data.Usage.CompletionTokens > usage.CompletionTokens {
			usage.CompletionTokens = newResponse.Data.Usage.CompletionTokens
		}
		usage.TotalTokens = usage.PromptTokens + usage.CompletionTokens
		response := streamResponse2OpenAI(&newResponse)
		if response == nil {
			continue
		}
		err = render.ObjectData(c, response)
		if err != nil {
			logger.SysError(err.Error())
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}

	render.Done(c)

	err := resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, &usage
}

func responseSensenova2OpenAI(response *ChatResponse) *openai.TextResponse {
	choices := make([]openai.TextResponseChoice, len(response.Data.Choices))
	for i, choice := range response.Data.Choices {
		choices[i] = openai.TextResponseChoice{
			Index: i,
			Message: model.Message{
				Role:             choice.Role,
				Content:          choice.Message,
				ReasoningContent: choice.ReasoningContent,
			},
			FinishReason: choice.FinishReason,
		}
	}
	newResponse := &openai.TextResponse{
		Id:      response.Data.ID,
		Choices: choices,
		Usage: model.Usage{
			PromptTokens:     response.Data.Usage.PromptTokens,
			CompletionTokens: response.Data.Usage.CompletionTokens,
			TotalTokens:      response.Data.Usage.TotalTokens,
		},
	}
	return newResponse
}

// func streamResponse2OpenAI(response *StreamResponse) *openai.StreamResponse {
func streamResponse2OpenAI(newResponse *StreamResponse) *openai.ChatCompletionsStreamResponse {
	if len(newResponse.Data.Choices) == 0 {
		return nil
	}
	newChoice := newResponse.Data.Choices[0]
	var choice openai.ChatCompletionsStreamResponseChoice
	choice.Delta = model.Message{
		Role:             newChoice.Role,
		Content:          newChoice.Delta,
		ReasoningContent: newChoice.ReasoningContent,
	}
	if newChoice.FinishReason != "null" {
		finishReason := newChoice.FinishReason
		choice.FinishReason = &finishReason
	}
	response := openai.ChatCompletionsStreamResponse{
		Id:      newResponse.Data.ID,
		Object:  "chat.completion.chunk",
		Created: helper.GetTimestamp(),
		Choices: []openai.ChatCompletionsStreamResponseChoice{choice},
	}
	return &response
}
