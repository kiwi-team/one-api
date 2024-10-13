package bfl

type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	UserId string `json:"user_id,omitempty"`
}

type TaskResponse struct {
	Id     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`
	Result struct {
		Sample string `json:"sample,omitempty"`
		Prompt string `json:"prompt,omitempty"`
	} `json:"result,omitempty"`
}
