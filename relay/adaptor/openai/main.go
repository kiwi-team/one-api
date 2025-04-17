package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/contentcheck"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/render"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/conv"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay/model"
	"github.com/songquanpeng/one-api/relay/relaymode"
)

const (
	dataPrefix       = "data: "
	done             = "[DONE]"
	dataPrefixLength = len(dataPrefix)
)

func StreamHandler(c *gin.Context, resp *http.Response, relayMode int) (*model.ErrorWithStatusCode, string, *model.Usage) {
	responseText := ""
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)
	var usage *model.Usage

	common.SetEventStreamHeaders(c)

	doneRendered := false
	preStr := ""
	content := ""
	contentCheckResult := ""
	pass := false
	for scanner.Scan() {
		data := scanner.Text()
		if len(data) < dataPrefixLength { // ignore blank line or wrong format
			continue
		}
		if data[:dataPrefixLength] != dataPrefix && data[:dataPrefixLength] != done {
			continue
		}
		if strings.HasPrefix(data[dataPrefixLength:], done) {
			render.StringData(c, data)
			doneRendered = true
			continue
		}
		switch relayMode {
		case relaymode.ChatCompletions:
			var streamResponse ChatCompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				render.StringData(c, data) // if error happened, pass the data to client
				continue                   // just ignore the error
			}
			if len(streamResponse.Choices) == 0 && streamResponse.Usage == nil {
				// but for empty choice and no usage, we should not pass it to client, this is for azure
				continue // just ignore empty choice
			}
			render.StringData(c, data)
			if config.BaiduAKAndSK != "" {
				for _, choice := range streamResponse.Choices {
					// 调用百度的安全审查接口，每一个百个字符，查一下。
					tmp := conv.AsString(choice.Delta.Content)
					responseText += tmp
					content = content + regexp.MustCompile(`\s+`).ReplaceAllString(tmp, "")
					// utf8字符串长度>=100，则截断
					if utf8.RuneCountInString(content) >= config.ContentCheckPerLength {
						//go func() {
						pass, contentCheckResult = contentCheck(preStr, content)
						if !pass {
							render.BadRequest(c, contentCheckResult)
							return nil, "", nil
						}
						preStr = content
						content = ""
					}
				}

			}
			if streamResponse.Usage != nil {
				usage = streamResponse.Usage
			}
		case relaymode.Completions:
			render.StringData(c, data)
			var streamResponse CompletionsStreamResponse
			err := json.Unmarshal([]byte(data[dataPrefixLength:]), &streamResponse)
			if err != nil {
				logger.SysError("error unmarshalling stream response: " + err.Error())
				continue
			}
			for _, choice := range streamResponse.Choices {
				responseText += choice.Text
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logger.SysError("error reading stream: " + err.Error())
	}
	if utf8.RuneCountInString(content) > 0 {
		pass, contentCheckResult := contentCheck(preStr, content)
		if !pass {
			render.BadRequest(c, contentCheckResult)
			return nil, "", nil
		}

	}

	if !doneRendered {
		render.Done(c)
	}

	err := resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), "", nil
	}

	return nil, responseText, usage
}

func contentCheck(preStr string, content string) (bool, string) {
	pass, contentCheckResult, err := contentcheck.GetAdaptor(contentcheck.Keyword).CheckContent(helper.LastNChars(preStr, config.ContentCheckPrefixLength)+content, 0)
	if err != nil {
		logger.SysError("error keyword checking content: " + err.Error())
	}
	if !pass {
		// todo 调用 gpt-4o-mini 判断是否泄漏模型身份
		return pass, contentCheckResult
	}

	pass, contentCheckResult, err = contentcheck.GetAdaptor(contentcheck.Baidu).CheckContent(helper.LastNChars(preStr, config.ContentCheckPrefixLength)+content, 0)
	if err != nil {
		logger.SysError("error baidu checking content: " + err.Error())
	}
	if !pass {
		return pass, contentCheckResult
	}
	return true, ""
}

func Handler(c *gin.Context, resp *http.Response, promptTokens int, modelName string) (*model.ErrorWithStatusCode, *model.Usage) {
	var textResponse SlimTextResponse
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return ErrorWrapper(err, "read_response_body_failed", http.StatusInternalServerError), nil
	}
	err = resp.Body.Close()
	if err != nil {
		return ErrorWrapper(err, "close_response_body_failed", http.StatusInternalServerError), nil
	}
	err = json.Unmarshal(responseBody, &textResponse)
	isGLMZ1 := strings.HasPrefix(modelName, "glm-z1")
	if textResponse.Choices != nil && len(textResponse.Choices) > 0 && isGLMZ1 {
		content := textResponse.Choices[0].Message.StringContent()
		// Extract content between <think> tags
		/*
			(?s) 是正则表达式的 "single line" 模式标志 它让 . 能够匹配任何字符，包括换行符 \n 这样就能正确匹配多行的 <think> 内容了
		*/
		thinkRegex := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
		matches := thinkRegex.FindStringSubmatch(content)
		if len(matches) > 1 {
			// Set the reasoning content
			textResponse.Choices[0].Message.ReasoningContent = matches[1]
			// Remove the <think> section from the content
			content = thinkRegex.ReplaceAllString(content, "")
		}
		// Set the cleaned content
		textResponse.Choices[0].Message.Content = content
	}
	if err != nil {
		return ErrorWrapper(err, "unmarshal_response_body_failed", http.StatusInternalServerError), nil
	}
	if textResponse.Error.Type != "" {
		return &model.ErrorWithStatusCode{
			Error:      textResponse.Error,
			StatusCode: resp.StatusCode,
		}, nil
	}

	// Reset response body
	if isGLMZ1 {
		jsonResponse, _ := json.Marshal(textResponse)
		resp.Body = io.NopCloser(bytes.NewBuffer(jsonResponse))
	} else {
		resp.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}

	// We shouldn't set the header before we parse the response body, because the parse part may fail.
	// And then we will have to send an error response, but in this case, the header has already been set.
	// So the HTTPClient will be confused by the response.
	// For example, Postman will report error, and we cannot check the response at all.
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

	if textResponse.Usage.TotalTokens == 0 || (textResponse.Usage.PromptTokens == 0 && textResponse.Usage.CompletionTokens == 0 && textResponse.Usage.InputTokens == 0 && textResponse.Usage.OutputTokens == 0) {
		completionTokens := 0
		for _, choice := range textResponse.Choices {
			completionTokens += CountTokenText(choice.Message.StringContent(), modelName)
		}
		textResponse.Usage = model.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		}
	}
	return nil, &textResponse.Usage
}
