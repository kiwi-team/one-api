package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func RelayResponsesHelper(c *gin.Context, relayMode int) *relaymodel.ErrorWithStatusCode {

	ctx := c.Request.Context()
	meta := meta.GetByContext(c)
	// get & validate textRequest
	textRequest, err := getAndValidateNewModelRequest(c, meta.Mode)
	// 序列化测试
	//res1, _ := json.MarshalIndent(textRequest, "", "  ")
	//fmt.Println("\n Serialized case1:  ", string(res1))
	if err != nil {
		logger.Errorf(ctx, "getAndValidateNewModelRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_reponses_request", http.StatusBadRequest)
	}
	meta.IsStream = textRequest.Stream

	// map model name
	meta.OriginModelName = textRequest.Model
	textRequest.Model, _ = getMappedModelName(textRequest.Model, meta.ModelMapping)
	meta.ActualModelName = textRequest.Model
	// get model ratio & group ratio
	modelRatio := billingratio.GetModelRatio(textRequest.Model, meta.ChannelType)
	groupRatio := billingratio.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio
	// pre-consume quota
	promptTokens := 0
	meta.PromptTokens = promptTokens
	ctx, preConsumedQuota, bizErr := preConsumeQuota(ctx, &relaymodel.GeneralOpenAIRequest{}, promptTokens, ratio, meta)
	if bizErr != nil {
		logger.Warnf(ctx, "preConsumeQuota failed: %+v", *bizErr)
		return bizErr
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// get request body
	// requestBody, err := getRequestBody(c, meta, textRequest, adaptor)
	// if err != nil {
	// 	return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	// }
	// Convert textRequest to JSON string
	textRequestBytes, _ := json.Marshal(textRequest)
	requestBodyContent := string(textRequestBytes)
	requestBody := bytes.NewBuffer(textRequestBytes)

	// do request
	resp, err := adaptor.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if isErrorHappened(meta, resp) {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return RelayErrorHandler(resp)
	}

	var responseBodyBuf bytes.Buffer

	// Create TeeReader to copy the response body and assign it back to resp.Body
	tee := io.TeeReader(resp.Body, &responseBodyBuf)
	resp.Body = io.NopCloser(tee)

	// do response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}

	responseBodyBytes, _ := io.ReadAll(&responseBodyBuf)
	responseBodyContent := string(responseBodyBytes)

	// post-consume quota
	go postConsumeQuota(ctx, usage, meta, textRequest.Model, ratio, preConsumedQuota, modelRatio, groupRatio, requestBodyContent, responseBodyContent)
	return nil
}
