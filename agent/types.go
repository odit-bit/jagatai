package agent

import (
	"context"
	"time"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	ToolCallID string     `json:"tool_call_id"`
	Toolcalls  []ToolCall `json:"tool_calls"`
}

// ChatCompletionRequest
type CCReq struct {
	Model      string    `json:"model"`
	Messages   []Message `json:"messages"`
	Stream     bool      `json:"stream"`
	Think      bool      `json:"think"`
	Tools      []Tool    `json:"tools"`
	ToolChoice string    `json:"tool_choice"`
}

// ChatCompletionResponse
type CCRes struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Object  string    `json:"object"`
	Created Timestamp `json:"created"`
	Choices []Choice  `json:"choices"`
	Usage   usage     `json:"usage"`
}

func (res *CCRes) IsToolCall() bool {
	if len(res.Choices) > 0 {
		if len(res.Choices[0].Message.Toolcalls) > 0 {
			return true
		}
	}
	return false
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
	Delta        Message `json:"delta"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Models struct {
	Data []Model `json:"data"`
}

type Model struct {
	ID      string    `json:"model"`
	Object  string    `json:"object"`
	Created Timestamp `json:"created"`
	OwnedBy string    `json:"owned_by"`
}

type Timestamp int64

func (tm *Timestamp) String() string {
	return time.Unix(int64(*tm), 0).String()
}

func (tm *Timestamp) Time() time.Time {
	return time.Unix(int64(*tm), 0)
}

// Remote backend for model
type Provider interface {
	// Chat(ctx context.Context, req CCReq, fn func(res CCRes) error) error
	Chat(ctx context.Context, req CCReq) (*CCRes, error)
}

type ToolProviders interface {
	Invoke(ctx context.Context, tc ToolCall) (string, error)
	ToSlice() []Tool
}
