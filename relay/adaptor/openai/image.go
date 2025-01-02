package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/middleware"
	"github.com/songquanpeng/one-api/relay/model"
)

func ImageHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	var imageResponse ImageResponse
	responseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	// 如果是LongCat-T2I-Medium 需要把响应改成文生图的格式
	if middleware.GetModel(c) == "LongCat-T2I-Medium" {
		textResponse := TextResponse{}
		err = json.Unmarshal(responseBody, &textResponse)
		if err != nil {
			return ErrorWrapper(err, "longCat-t2i-unmarshal_response_body_failed", http.StatusInternalServerError), nil
		}
		imageResponse := ImageResponse{
			Created: helper.GetTimestamp(),
		}
		for _, choice := range textResponse.Choices {
			imageResponse.Data = append(imageResponse.Data, ImageData{
				Url: choice.Message.StringContent(),
			})
		}
		responseBody, err = json.Marshal(imageResponse)
		if err != nil {
			return ErrorWrapper(err, "longCat-t2i-marshal_response_body_failed", http.StatusInternalServerError), nil
		}
	} else {
		err = json.Unmarshal(responseBody, &imageResponse)
	}
	if err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))

	for k, v := range resp.Header {
		c.Writer.Header().Set(k, v[0])
	}
	c.Writer.WriteHeader(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		return ErrorWrapper(err, "copy_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	return nil, nil
}
