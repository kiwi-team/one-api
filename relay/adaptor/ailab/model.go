package ailab

type ImageSubmitTaskRequest struct {
	Model  string `json:"model,omitempty"`
	Prompt string `json:"prompt"`
	Size   string `json:"size,omitempty" default:"1:1"`
	Type   int    `json:"type"`
}

type ImageSubmitTaskResponse struct {
	Code    int64  `json:"code"`
	Msg     string `json:"msg"`
	TraceID string `json:"traceId"`
	Data    struct {
		SessionID string `json:"session_id"`
		TaskID    int    `json:"task_id"`
	} `json:"data"`
}

type ImageTaskData struct {
	SessionID     string  `json:"session_id"`
	TaskID        int     `json:"task_id"`
	State         int     `json:"state"`
	QueuePosition int     `json:"queue_position"`
	Progress      float64 `json:"progress"`
	Output        string  `json:"output"`
	FeedbackFlag  int     `json:"feedback_flag"`
}

type ImageQueryTaskResponse struct {
	Code    int64         `json:"code"`
	Msg     string        `json:"msg"`
	TraceID string        `json:"traceId"`
	Data    ImageTaskData `json:"data"`
}
