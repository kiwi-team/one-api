package openai

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ChatCompletionRequest represents the request body for chat completion API
type ChatCompletionRequest struct {
	// Input can be either a string or an array of messages
	Input *InputData `json:"input,omitempty"`
	Model string     `json:"model"`

	Include            []string          `json:"include,omitempty"` // 取值范围['file_search_call.results', 'message.input_image.image_url', 'computer_call_output.output.image_url']
	Instructions       string            `json:"instructions,omitempty"`
	MaxOutputTokens    int               `json:"max_output_tokens,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	ParallelToolCalls  bool              `json:"parallel_tool_calls,omitempty"`
	PreviousResponseID string            `json:"previous_response_id,omitempty"`
	Reasoning          *Reasoning        `json:"reasoning,omitempty"`
	Store              bool              `json:"store,omitempty"`
	Stream             bool              `json:"stream,omitempty"`
	Temperature        float64           `json:"temperature,omitempty"`
	TopP               float64           `json:"top_p,omitempty"`
	Text               any               `json:"text,omitempty"`
	ToolChoice         *ToolChoice       `json:"tool_choice,omitempty"`
	Truncation         string            `json:"truncation,omitempty"`
	User               string            `json:"user,omitempty"`
	Tools              *Tools            `json:"tools,omitempty"`
}

// 定义输入数据类型
type InputData struct {
	Text     *string // 直接文本输入
	Messages []any   // 消息列表输入
}

// 自定义JSON解析
func (i *InputData) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		i.Text = &text
		return nil
	}

	// 尝试解析为消息数组
	var messages []any
	if err := json.Unmarshal(data, &messages); err == nil {
		i.Messages = messages
		return nil
	}

	return errors.New("input must be string or message array")
}

// 自定义JSON序列化
func (i *InputData) MarshalJSON() ([]byte, error) {
	switch {
	case i.Text != nil:
		return json.Marshal(i.Text)
	case i.Messages != nil:
		return json.Marshal(i.Messages)
	default:
		return json.Marshal(nil)
	}
}

type InputDataItem struct {
	InputMessage   *InputMessage
	InputItem      *InputItem
	InputReference *InputReference
}

type InputItem struct{}

type InputMessage struct {
	Content *InputMessageContent `json:"content"`        // Can be string or []InputItem
	Role    string               `json:"role"`           // One of user, assistant, system, or developer
	Type    string               `json:"type,omitempty"` // The type of the message input. Always message.
}

// Content 表示可以是字符串或输入项列表的联合类型
type InputMessageContent struct {
	Value any
}

// InputItem 表示单个输入项的通用结构
type InputMessageItem struct {
	Type string `json:"type"`

	// 文本输入字段
	Text string `json:"text,omitempty"`

	// 图片输入字段
	Detail   string `json:"detail,omitempty"`
	FileID   string `json:"file_id,omitempty"`
	ImageURL string `json:"image_url,omitempty"`

	// 文件输入字段
	FileData string `json:"file_data,omitempty"`
	Filename string `json:"filename,omitempty"`
}

// 自定义JSON序列化
func (c *InputMessageContent) MarshalJSON() ([]byte, error) {
	switch v := c.Value.(type) {
	case string:
		return json.Marshal(v)
	case []InputMessageItem:
		return json.Marshal(v)
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}

// 自定义JSON反序列化
func (c *InputMessageContent) UnmarshalJSON(data []byte) error {
	// 尝试解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.Value = s
		return nil
	}

	// 尝试解析为输入项列表
	var items []InputMessageItem
	if err := json.Unmarshal(data, &items); err == nil {
		c.Value = items
		return nil
	}

	return fmt.Errorf("content must be string or array")
}

type InputReference struct {
	ID   string `json:"id"`
	Type string `json:"type" default:"item_reference"`
}

type Reasoning struct {
	Effort          string `json:"effort,omitempty" default:"medium"` // Currently supported values are low, medium, and high
	GenerateSummary string `json:"generate_summary,omitempty"`        // One of concise or detailed

}

// ToolChoice represents how the model should select which tool to use

// 总类型定义
type ToolChoice struct {
	StringVal  *string             // 字符串类型 ("none", "auto", "required")
	HostedTool *HostedTool         // Hosted Tool 类型
	FuncTool   *FunctionToolChioce // Function Tool 类型
}

// Hosted Tool 结构 (type=file_search/web_search_preview/computer_use_preview)
type HostedTool struct {
	Type string `json:"type"` // 必须字段
}

// Function Tool 结构 (type=function)
type FunctionToolChioce struct {
	Type string `json:"type"`     // 必须为 "function"
	Name string `json:"function"` // 必须字段（文档中实际结构需要确认）
}

// 实现自定义 JSON 解析
func (tc *ToolChoice) UnmarshalJSON(data []byte) error {
	// 先尝试解析为字符串
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		valid := map[string]bool{"none": true, "auto": true, "required": true}
		if !valid[s] {
			return errors.New("invalid string value for tool_choice")
		}
		tc.StringVal = &s
		return nil
	}

	// 尝试解析为对象
	var temp struct {
		Type     string `json:"type"`
		Function struct {
			Name string `json:"name"`
		} `json:"function"`
	}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	switch temp.Type {
	case "function":
		if temp.Function.Name == "" {
			return errors.New("missing function name")
		}
		tc.FuncTool = &FunctionToolChioce{
			Type: "function",
			Name: temp.Function.Name,
		}
	case "file_search", "web_search_preview", "computer_use_preview":
		tc.HostedTool = &HostedTool{
			Type: temp.Type,
		}
	default:
		return fmt.Errorf("invalid object type: %s", temp.Type)
	}
	return nil
}

// 实现 JSON 序列化
func (tc ToolChoice) MarshalJSON() ([]byte, error) {
	switch {
	case tc.StringVal != nil:
		return json.Marshal(tc.StringVal)
	case tc.HostedTool != nil:
		return json.Marshal(tc.HostedTool)
	case tc.FuncTool != nil:
		return json.Marshal(struct {
			Type     string `json:"type"`
			Function struct {
				Name string `json:"name"`
			} `json:"function"`
		}{
			Type: "function",
			Function: struct {
				Name string `json:"name"`
			}{Name: tc.FuncTool.Name},
		})
	}
	return json.Marshal(nil)
}

// FileSearchTool 文件搜索工具
type FileSearchTool struct {
	Type           string          `json:"type"`
	VectorStoreIDs []string        `json:"vector_store_ids"`
	Filters        []Filter        `json:"filters,omitempty"`
	MaxNumResults  int             `json:"max_num_results,omitempty"`
	RankingOptions *RankingOptions `json:"ranking_options,omitempty"`
}

// RankingOptions 文件搜索的排名选项
type RankingOptions struct {
	Ranker         string  `json:"ranker,omitempty"`
	ScoreThreshold float64 `json:"score_threshold,omitempty"`
}

// Filter 过滤器，可以是比较过滤器或复合过滤器
type Filter struct {
	// 比较过滤器字段
	Key      string `json:"key,omitempty"`
	Operator string `json:"operator,omitempty"`
	Value    any    `json:"value,omitempty"`
	// 复合过滤器字段
	FilterType string   `json:"type,omitempty"` // "and" 或 "or"
	Filters    []Filter `json:"filters,omitempty"`
}

// FunctionTool 自定义函数工具
type FunctionTool struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Parameters  any    `json:"parameters"`
	Strict      bool   `json:"strict"`
	Description string `json:"description,omitempty"`
}

// WebSearchTool 网页搜索工具
type WebSearchTool struct {
	Type              string        `json:"type"`
	SearchContextSize string        `json:"search_context_size,omitempty"`
	UserLocation      *UserLocation `json:"user_location,omitempty"`
}

// UserLocation 用户近似位置
type UserLocation struct {
	Type     string `json:"type"`
	City     string `json:"city,omitempty"`
	Country  string `json:"country,omitempty"`
	Region   string `json:"region,omitempty"`
	Timezone string `json:"timezone,omitempty"`
}

// ComputerUseTool 计算机使用工具
type ComputerUseTool struct {
	Type          string  `json:"type"`
	DisplayHeight float64 `json:"display_height"`
	DisplayWidth  float64 `json:"display_width"`
	Environment   string  `json:"environment"`
}

// Tools 定义工具数组，每个元素可以是不同种类的工具
type Tools []Tool

// Tool 作为接口类型，用于统一处理各种工具
type Tool any

// UnmarshalJSON 自定义反序列化方法
func (t *Tools) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*t = make(Tools, len(raw))
	for i, item := range raw {
		var base struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(item, &base); err != nil {
			return err
		}

		switch base.Type {
		case "file_search":
			var tool FileSearchTool
			if err := json.Unmarshal(item, &tool); err != nil {
				return err
			}
			(*t)[i] = tool
		case "function":
			var tool FunctionTool
			if err := json.Unmarshal(item, &tool); err != nil {
				return err
			}
			(*t)[i] = tool
		case "web_search_preview", "web_search_preview_2025_03_11":
			var tool WebSearchTool
			if err := json.Unmarshal(item, &tool); err != nil {
				return err
			}
			(*t)[i] = tool
		case "computer_use_preview":
			var tool ComputerUseTool
			if err := json.Unmarshal(item, &tool); err != nil {
				return err
			}
			(*t)[i] = tool
		default:
			return fmt.Errorf("未知工具类型: %s", base.Type)
		}
	}
	return nil
}
