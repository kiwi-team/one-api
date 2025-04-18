package client

import (
	"net/http"
	"time"

	"github.com/songquanpeng/one-api/common/logger"
)

// RetryConfig 定义重试配置
type RetryConfig struct {
	MaxRetries       int           // 最大重试次数
	RetryWaitTime    time.Duration // 重试等待时间
	MaxRetryWaitTime time.Duration // 最大重试等待时间
}

// DefaultRetryConfig 返回默认的重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:       3,
		RetryWaitTime:    time.Second,
		MaxRetryWaitTime: time.Second * 10,
	}
}

// RetryableHTTPClient 是一个支持重试的 HTTP 客户端
type RetryableHTTPClient struct {
	client *http.Client
	config *RetryConfig
}

// NewRetryableHTTPClient 创建一个新的可重试 HTTP 客户端
func NewRetryableHTTPClient(client *http.Client, config *RetryConfig) *RetryableHTTPClient {
	if config == nil {
		config = DefaultRetryConfig()
	}
	return &RetryableHTTPClient{
		client: client,
		config: config,
	}
}

// Do 执行 HTTP 请求，支持自动重试
func (c *RetryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	waitTime := c.config.RetryWaitTime

	for i := 0; i <= c.config.MaxRetries; i++ {
		if i > 0 {
			logger.SysLogf("Retrying request %s %s (attempt %d/%d)", req.Method, req.URL.String(), i, c.config.MaxRetries)
			time.Sleep(waitTime)
			// 增加等待时间，但不超过最大等待时间
			waitTime = time.Duration(float64(waitTime) * 1.5)
			if waitTime > c.config.MaxRetryWaitTime {
				waitTime = c.config.MaxRetryWaitTime
			}
		}

		resp, err = c.client.Do(req)
		if err == nil {
			// 检查响应状态码，只对服务器错误进行重试
			if resp.StatusCode < 500 {
				return resp, nil
			}
			resp.Body.Close()
		}

		// 如果是最后一次尝试，直接返回错误
		if i == c.config.MaxRetries {
			if err != nil {
				return nil, err
			}
			return resp, nil
		}
	}

	return resp, err
}

// Get 执行 GET 请求
func (c *RetryableHTTPClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

// Post 执行 POST 请求
func (c *RetryableHTTPClient) Post(url string, contentType string, body interface{}) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return c.Do(req)
}
