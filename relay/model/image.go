package model

import "encoding/json"

// https://platform.openai.com/docs/api-reference/images/createEdit
type ImageRequest struct {
	Model          string          `json:"model"`
	Prompt         string          `json:"prompt" binding:"required"`
	N              int             `json:"n,omitempty"`
	Size           string          `json:"size,omitempty"`
	Quality        string          `json:"quality,omitempty"`
	ResponseFormat string          `json:"response_format,omitempty"`
	Style          string          `json:"style,omitempty"`
	User           string          `json:"user,omitempty"`
	ExtraFields    json.RawMessage `json:"extra_fields,omitempty"`
	Background     string          `json:"background,omitempty"`
	Moderation     string          `json:"moderation,omitempty"`
	OutputFormat   string          `json:"output_format,omitempty"`
}
