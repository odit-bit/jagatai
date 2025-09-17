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

type Message struct {
	Role  Role
	Parts []*Part
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

type Blob struct {
	//raw bytes
	Bytes []byte
	//IANA standart type
	Mime string
}

// ChatCompletionResponse present result receive from provider
type CCRes struct {
	ID      string    `json:"id"`
	Model   string    `json:"model"`
	Object  string    `json:"object"`
	Created time.Time `json:"created"`
	Choices []Choice  `json:"choices"`
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

type Choice struct {
	Index        int `json:"index"`
	Text         string
	ToolCalls    []*ToolCall `json:"tool_calls,omitempty"`
	FinishReason string      `json:"finish_reason"`
	Delta        Message     `json:"delta"`
}

type Usage struct {
	PromptTokens     int32 `json:"prompt_tokens"`
	CompletionTokens int32 `json:"completion_tokens"`
	TotalTokens      int32 `json:"total_tokens"`
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

// helper

func NewTextMessage(role Role, text string) *Message {
	m := &Message{
		Role: role,
		Parts: []*Part{
			{Text: text},
		},
	}
	return m
}

func NewBlobMessage(role Role, b []byte, mime string) *Message {
	m := Message{
		Role: role,
		Parts: []*Part{
			{
				Blob: &Blob{
					Bytes: b,
					Mime:  mime,
				},
			},
		},
	}
	return &m
}
