package openai

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/alibailian"
	"github.com/songquanpeng/one-api/relay/adaptor/baidu2"
	"github.com/songquanpeng/one-api/relay/adaptor/baiduv2"
	"github.com/songquanpeng/one-api/relay/adaptor/doubao"
	"github.com/songquanpeng/one-api/relay/adaptor/friday"
	"github.com/songquanpeng/one-api/relay/adaptor/geminiv2"
	"github.com/songquanpeng/one-api/relay/adaptor/minimax"
	"github.com/songquanpeng/one-api/relay/adaptor/newaliyun"
	"github.com/songquanpeng/one-api/relay/adaptor/novita"
	"github.com/songquanpeng/one-api/relay/channeltype"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Adaptor struct {
	ChannelType int
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.ChannelType = meta.ChannelType
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.ChannelType {
	case channeltype.Azure:
		if meta.Mode == relaymode.ImagesGenerations {
			// https://learn.microsoft.com/en-us/azure/ai-services/openai/dall-e-quickstart?tabs=dalle3%2Ccommand-line&pivots=rest-api
			// https://{resource_name}.openai.azure.com/openai/deployments/dall-e-3/images/generations?api-version=2024-03-01-preview
			fullRequestURL := fmt.Sprintf("%s/openai/deployments/%s/images/generations?api-version=%s", meta.BaseURL, meta.ActualModelName, meta.Config.APIVersion)
			return fullRequestURL, nil
		}

		// https://learn.microsoft.com/en-us/azure/cognitive-services/openai/chatgpt-quickstart?pivots=rest-api&tabs=command-line#rest-api
		requestURL := strings.Split(meta.RequestURLPath, "?")[0]
		requestURL = fmt.Sprintf("%s?api-version=%s", requestURL, meta.Config.APIVersion)
		task := strings.TrimPrefix(requestURL, "/v1/")
		model_ := meta.ActualModelName
		model_ = strings.Replace(model_, ".", "", -1)
		//https://github.com/songquanpeng/one-api/issues/1191
		// {your endpoint}/openai/deployments/{your azure_model}/chat/completions?api-version={api_version}
		requestURL = fmt.Sprintf("/openai/deployments/%s/%s", model_, task)
		return GetFullRequestURL(meta.BaseURL, requestURL, meta.ChannelType), nil
	case channeltype.Minimax:
		return minimax.GetRequestURL(meta)
	case channeltype.Doubao:
		return doubao.GetRequestURL(meta)
	case channeltype.Novita:
		return novita.GetRequestURL(meta)
	case channeltype.Friday:
		return friday.GetRequestURL(meta)
	case channeltype.Newaliyun:
		return newaliyun.GetRequestURL(meta)
	case channeltype.Baidu2:
		return baidu2.GetRequestURL(meta)
	case channeltype.BaiduV2:
		return baiduv2.GetRequestURL(meta)
	case channeltype.AliBailian:
		return alibailian.GetRequestURL(meta)
	case channeltype.GeminiOpenAICompatible:
		return geminiv2.GetRequestURL(meta)
	default:
		return GetFullRequestURL(meta.BaseURL, meta.RequestURLPath, meta.ChannelType), nil
	}
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	if meta.ChannelType == channeltype.Azure {
		req.Header.Set("api-key", meta.APIKey)
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	if meta.ChannelType == channeltype.OpenRouter {
		req.Header.Set("HTTP-Referer", "https://github.com/songquanpeng/one-api")
		req.Header.Set("X-Title", "One API")
	}
	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if request.Stream || slices.Contains(alibailian.StreamOnlyModelList, request.Model) {
		request.Stream = true
		// always return usage in stream mode
		if request.StreamOptions == nil {
			request.StreamOptions = &model.StreamOptions{}
		}
		request.StreamOptions.IncludeUsage = true
	}
	if a.ChannelType == channeltype.Baidu2 {
		request.Model = strings.ToLower(request.Model)
	}
	// 走panda渠道的时候,如果请求是 claude-3-7-sonnet-20250219 并且开启了 thinking，则把模型名改为 claude-3-7-sonnet-20250219#thinking
	if a.ChannelType == channeltype.Panda && strings.HasPrefix(request.Model, "claude-") && request.Thinking != nil && request.Thinking.Type == "enabled" {
		//request.Model = request.Model + "#thinking"
		if request.Thinking.BudgetTokens > 0 {
			request.Reasoning = &model.Reasoning{
				MaxTokens: &request.Thinking.BudgetTokens,
			}
		} else {
			request.Reasoning = &model.Reasoning{
				Effort: "medium",
			}
		}
		if request.Thinking.BudgetTokens > request.MaxTokens {
			request.MaxTokens = request.Thinking.BudgetTokens + 1
		}
	}
	// 走panda的渠道， gemini-2.5-flash-preview-04-17 开启 thinking 要按照panda的格式来传参数
	if a.ChannelType == channeltype.Panda && strings.HasPrefix(request.Model, "gemini-2.5-flash-preview-04-17") && request.Thinking != nil {
		if request.Thinking.Type == "enabled" && request.Thinking.BudgetTokens >= 0 {
			request.Reasoning = &model.Reasoning{
				MaxTokens: &request.Thinking.BudgetTokens,
			}
		} else if request.Thinking.Type != "enabled" {
			maxTokens := 0
			request.Reasoning = &model.Reasoning{
				MaxTokens: &maxTokens,
			}
		}
	}
	if strings.HasPrefix(request.Model, "gemini") {
		//  兼容爱果果/panda的请求把 audio_url, video_url 转换为 image_url
		if a.ChannelType == channeltype.Aiguoguo || a.ChannelType == channeltype.Panda {
			newMessages := make([]model.Message, 0, len(request.Messages))
			for _, message := range request.Messages {
				newContentArr := make([]model.MessageContent, 0)
				arr := message.ParseContent()
				for _, content := range arr {
					var newContent model.MessageContent
					switch content.Type {
					case model.ContentTypeAudioURL:
						newContent = model.MessageContent{
							Type: model.ContentTypeImageURL,
							ImageURL: &model.ImageURL{
								Url: content.AudioURL.Url,
							},
						}
					case model.ContentTypeVideoURL:
						newContent = model.MessageContent{
							Type: model.ContentTypeImageURL,
							ImageURL: &model.ImageURL{
								Url: content.VideoURL.Url,
							},
						}
					default:
						newContent = content
					}
					newContentArr = append(newContentArr, newContent)
				}
				newMessage := message
				newMessage.Content = newContentArr
				newMessages = append(newMessages, newMessage)
			}
			request.Messages = newMessages
		}
	}
	if a.ChannelType == channeltype.Aiguoguo {
		request.Thinking.IncludeThinking = false // 响应体内容是否和思考内容合并, true则二者合并在content中, false或者不传则默认在
	}
	return request, nil
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if request.Model == "LongCat-T2I-Medium" {
		newRequest := TextRequest{
			Model: request.Model,
			Messages: []model.Message{
				{
					Role:    "user",
					Content: request.Prompt,
				},
			},
		}
		return newRequest, nil
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	if meta.IsStream {
		var responseText string
		err, responseText, usage = StreamHandler(c, resp, meta.Mode)
		if usage == nil || usage.TotalTokens == 0 {
			usage = ResponseText2Usage(responseText, meta.ActualModelName, meta.PromptTokens)
		}
		// panda data: {"id":"chatcmpl-6887ccb0136e47ab9dc74238955a7d94","object":"chat.completion.chunk","created":1745139709,"model":"claude-3-7-sonnet-20250219","choices":[],"usage":{"prompt_tokens":0,"completion_tokens":1439,"total_tokens":1439,"prompt_tokens_details":{},"completion_tokens_details":{"reasoning_tokens":0,"accepted_prediction_tokens":0,"rejected_prediction_tokens":0}}}
		// 直接按照第三方的返回来记录，便于追溯
		//if usage.TotalTokens != 0 && usage.PromptTokens == 0 { // some channels don't return prompt tokens & completion tokens
		//usage.PromptTokens = meta.PromptTokens
		//usage.CompletionTokens = usage.TotalTokens - meta.PromptTokens
		//}
	} else {
		switch meta.Mode {
		case relaymode.ImagesGenerations:
			err, _ = ImageHandler(c, resp)
		default:
			// 对于只能流式返回的模型，拼接数据，整体返回
			if slices.Contains(alibailian.StreamOnlyModelList, meta.ActualModelName) {
				err, usage = OnlyStreamModelHandler(c, resp, meta.ActualModelName)
			} else {
				err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
			}
		}
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	_, modelList := GetCompatibleChannelMeta(a.ChannelType)
	return modelList
}

func (a *Adaptor) GetChannelName() string {
	channelName, _ := GetCompatibleChannelMeta(a.ChannelType)
	return channelName
}
