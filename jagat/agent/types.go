package agent

import (
	"strings"
	"time"
)

type Role string

const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
	RoleSystem    Role = "system"
	RoleTool      Role = "tool"
)

const (
	TextPart     = "text"
	BlobPart     = "blob"
	ToolCallPart = "toolcall"
	ToolRespPart = "toolresp"
)

type Messages []*Message

// ChatCompletionRequest use for communicating with provider
type CCReq struct {
	// Model      string
	Messages   []*Message
	Stream     bool
	Think      bool
	Tools      []Tool
	ToolChoice string
}

// represent single message in conversation or history.
// it compose from multiple part of diferrent kind.
type Message struct {
	// the role of sender message
	Role Role
	// each part only have one field filled and should turn error if more.
	Parts []*Part
}

func (m *Message) ToolCall() (*ToolCall, bool) {
	for _, p := range m.Parts {
		if p.Toolcall != nil {
			return p.Toolcall, true
		}
	}
	return nil, false
}

func (m *Message) Text() string {
	texts := []string{}
	for _, p := range m.Parts {
		if p.Text != "" {
			texts = append(texts, p.Text)
		}
	}
	return strings.Join(texts, "")
}

func NewPartToolCall(name, arg string) *Part {
	return &Part{
		Toolcall: &ToolCall{
			ID: "",
		},
	}
}

type Part struct {
	Text         string
	Blob         *Blob
	Toolcall     *ToolCall
	ToolResponse *ToolResponse
}

// represent inline raw binary data in message.
type Blob struct {
	//raw bytes
	Bytes []byte
	//IANA standart type (jpeg, pdf)
	Mime string
}

// ChatCompletionResponse present result receive from provider
type CCRes struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Object  string    `json:"object"`
	Created time.Time `json:"created"`
	//if supported and configured, the provider could response more than one choice, but in this implementation, agent always use choices[0]
	Choices []Choice `json:"choices"`
	Usage   Usage
}

func (res *CCRes) IsToolCall() ([]*ToolCall, bool) {
	if len(res.Choices) > 0 {
		tc := res.Choices[0].ToolCalls
		if len(tc) > 0 {
			return tc, true
		}
	}
	return nil, false

}

// choice represent single response from provider.
type Choice struct {
	Index        int         `json:"index"`
	Text         string      `json:"text"`
	ToolCalls    []*ToolCall `json:"tool_calls,omitempty"`
	FinishReason string      `json:"finish_reason"`
	Delta        Message     `json:"delta"`
}

type Usage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
}
