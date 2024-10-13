package klingai

import (
	"time"
)

type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	UserId string `json:"user_id,omitempty"`
}

type tokenData struct {
	Token      string
	ExpiryTime time.Time
}

type TaskResponse struct {
	Code      int    `json:"code,omitempty"`
	Message   string `json:"message,omitempty"`
	RequestId string `json:"request_id,omitempty"`
	Data      struct {
		TaskId        string `json:"task_id,omitempty"`
		TaskStatus    string `json:"task_status,omitempty"`
		CreatedAt     int64  `json:"created_at,omitempty"`
		UpdatedAt     int64  `json:"updated_at,omitempty"`
		TaskStatusMsg string `json:"task_status_msg,omitempty"`
		TaskResult    struct {
			Images []struct {
				Index int    `json:"index,omitempty"`
				Url   string `json:"url,omitempty"`
			} `json:"images,omitempty"`
		} `json:"task_result,omitempty"`
	} `json:"data,omitempty"`
}
