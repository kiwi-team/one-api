package ai302

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

type Adaptor struct {
	meta *meta.Meta
}

func (a *Adaptor) Init(meta *meta.Meta) {
	a.meta = meta
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)
	req.Header.Set("Authorization", "Bearer "+meta.APIKey)
	req.Header.Set("Content-Type", "application/json")
	return nil
}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	switch meta.Mode {
	case relaymode.ChatCompletions:
		return fmt.Sprintf("%s/v1/chat/completions", meta.BaseURL), nil
	case relaymode.Embeddings:
		return fmt.Sprintf("%s/v1/embeddings", meta.BaseURL), nil
	case relaymode.ImagesGenerations:
		if meta.OriginModelName == "recraft-v3" {
			return fmt.Sprintf("%s/302/submit/recraft-v3", meta.BaseURL), nil
		}
	default:
	}
	return "", fmt.Errorf("unsupported relay mode %d for doubao", meta.Mode)
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	if request.Model == "recraft-v3" {
		newRequest := Recraft302ImageRequest{
			Prompt: request.Prompt,
			ImageSize: struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			}{
				Width:  1024,
				Height: 1024,
			},
			Style: request.Style,
		}
		return newRequest, nil
	}
	return request, nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	if request == nil {
		return nil, errors.New("request is nil")
	}
	return request, nil
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	switch meta.Mode {
	case relaymode.ImagesGenerations:
		if meta.OriginModelName == "recraft-v3" {
			err, usage = RecraftImageHandler(c, resp)
		} else {
			err, usage = openai.ImageHandler(c, resp)
		}
		return
	case relaymode.ChatCompletions:
		err, usage = openai.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	case relaymode.Embeddings:
		err, usage = openai.Handler(c, resp, meta.PromptTokens, meta.ActualModelName)
	}
	return
}

func (a *Adaptor) GetModelList() []string {
	return ModelList
}

func (a *Adaptor) GetChannelName() string {
	return "ai302"
}
