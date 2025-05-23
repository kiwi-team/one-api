package openai

import (
	"errors"
	"fmt"
	"io"
	"net/http"
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
	if request.Stream {
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
	if a.ChannelType == channeltype.Panda && strings.HasPrefix(request.Model, "claude-3-7-sonnet-20250219") && request.Thinking != nil && request.Thinking.Type == "enabled" {
		request.Model = request.Model + "#thinking"
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

/*
func (a *Adaptor) ConvertImageRequest111(c *gin.Context, info *relaycommon.RelayInfo, request dto.ImageRequest) (any, error) {
	switch info.RelayMode {
	case constant.RelayModeImagesEdits:

		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		writer.WriteField("model", request.Model)
		// 获取所有表单字段
		formData := c.Request.PostForm
		// 遍历表单字段并打印输出
		for key, values := range formData {
			if key == "model" {
				continue
			}
			for _, value := range values {
				writer.WriteField(key, value)
			}
		}

		// Parse the multipart form to handle both single image and multiple images
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB max memory
			return nil, errors.New("failed to parse multipart form")
		}

		if c.Request.MultipartForm != nil && c.Request.MultipartForm.File != nil {
			// Check if "image" field exists in any form, including array notation
			var imageFiles []*multipart.FileHeader
			var exists bool

			// First check for standard "image" field
			if imageFiles, exists = c.Request.MultipartForm.File["image"]; !exists || len(imageFiles) == 0 {
				// If not found, check for "image[]" field
				if imageFiles, exists = c.Request.MultipartForm.File["image[]"]; !exists || len(imageFiles) == 0 {
					// If still not found, iterate through all fields to find any that start with "image["
					foundArrayImages := false
					for fieldName, files := range c.Request.MultipartForm.File {
						if strings.HasPrefix(fieldName, "image[") && len(files) > 0 {
							foundArrayImages = true
							for _, file := range files {
								imageFiles = append(imageFiles, file)
							}
						}
					}

					// If no image fields found at all
					if !foundArrayImages && (len(imageFiles) == 0) {
						return nil, errors.New("image is required")
					}
				}
			}

			// Process all image files
			for i, fileHeader := range imageFiles {
				file, err := fileHeader.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open image file %d: %w", i, err)
				}
				defer file.Close()

				// If multiple images, use image[] as the field name
				fieldName := "image"
				if len(imageFiles) > 1 {
					fieldName = "image[]"
				}

				// Determine MIME type based on file extension
				mimeType := detectImageMimeType(fileHeader.Filename)

				// Create a form file with the appropriate content type
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, fieldName, fileHeader.Filename))
				h.Set("Content-Type", mimeType)

				part, err := writer.CreatePart(h)
				if err != nil {
					return nil, fmt.Errorf("create form part failed for image %d: %w", i, err)
				}

				if _, err := io.Copy(part, file); err != nil {
					return nil, fmt.Errorf("copy file failed for image %d: %w", i, err)
				}
			}

			// Handle mask file if present
			if maskFiles, exists := c.Request.MultipartForm.File["mask"]; exists && len(maskFiles) > 0 {
				maskFile, err := maskFiles[0].Open()
				if err != nil {
					return nil, errors.New("failed to open mask file")
				}
				defer maskFile.Close()

				// Determine MIME type for mask file
				mimeType := detectImageMimeType(maskFiles[0].Filename)

				// Create a form file with the appropriate content type
				h := make(textproto.MIMEHeader)
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="mask"; filename="%s"`, maskFiles[0].Filename))
				h.Set("Content-Type", mimeType)

				maskPart, err := writer.CreatePart(h)
				if err != nil {
					return nil, errors.New("create form file failed for mask")
				}

				if _, err := io.Copy(maskPart, maskFile); err != nil {
					return nil, errors.New("copy mask file failed")
				}
			}
		} else {
			return nil, errors.New("no multipart form data found")
		}

		// 关闭 multipart 编写器以设置分界线
		writer.Close()
		c.Request.Header.Set("Content-Type", writer.FormDataContentType())
		return bytes.NewReader(requestBody.Bytes()), nil

	default:
		return request, nil
	}
}
*/

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
			err, usage = Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
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
