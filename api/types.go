package api

import (
	"net/http"
	"time"

	"github.com/odit-bit/jagatai/jagat/agent"
)

// Request
type ChatRequest struct {
	Content []*Message `json:"content"`
}

type Message agent.Message

// type Message struct {
// 	Role  string  `json:"role"`
// 	Parts []*Part `json:"parts"`
// }

// type Part struct {
// 	Text string `json:"text"`
// 	Blob *Blob  `json:"blob"`
// }

// type Blob struct {
// 	Bytes []byte `json:"bytes"`
// 	Mime  string `json:"mime"`
// }

// Response
type ChatResponse struct {
	Created time.Time `json:"created"`
	Text    string    `json:"text"`
}

/* HELPER  */

func NewBlobMessage(Role string, b []byte, mimeType string) *Message {
	if mimeType == "" {
		mimeType = http.DetectContentType(b)
	}
	return &Message{
		Role: "user",
		Parts: []*agent.Part{
			newBlobPart(b, mimeType),
		},
	}
}

func NewTextMessage(role string, text string) *Message {
	return &Message{
		Role: agent.Role(role),
		Parts: []*agent.Part{
			NewTextPart(text),
		},
	}
}

func NewTextPart(text string) *agent.Part {
	return &agent.Part{
		Text: text,
	}
}

func newBlobPart(b []byte, mime string) *agent.Part {
	return &agent.Part{
		Blob: &agent.Blob{
			Bytes: b,
			Mime:  mime,
		},
	}
}
