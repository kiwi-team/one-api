package ai302

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/model"
)

func RecraftImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var imageResponse Recraft302ImageResponse
	var openaiImageResponse *openai.ImageResponse
	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return openai.ErrorWrapper(err, "read_302_recraft_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return openai.ErrorWrapper(err, "close_302_recraft_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &imageResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if len(imageResponse.Images) < 1 {
		return openai.ErrorWrapper(errors.New("recraft-v3-images-length-error"), "recraft-v3_response_images_error", http.StatusInternalServerError), nil
	}
	openaiImageResponse = responseRecraft2OpenAI(&imageResponse)
	jsonResponse, err := json.Marshal(openaiImageResponse)
	if err != nil {
		return openai.ErrorWrapper(err, "marshal_response_body_failed", http.StatusInternalServerError), nil
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(resp.StatusCode)
	_, _ = c.Writer.Write(jsonResponse)

	return nil, nil
}

func responseRecraft2OpenAI(recraftResponse *Recraft302ImageResponse) *openai.ImageResponse {
	imageResponse := openai.ImageResponse{
		Created: helper.GetTimestamp(),
	}

	imageResponse.Data = append(imageResponse.Data, openai.ImageData{
		Url:           recraftResponse.Images[0].URL,
		B64Json:       "",
		RevisedPrompt: "",
	})

	return &imageResponse
}
