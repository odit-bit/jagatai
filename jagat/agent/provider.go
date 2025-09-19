package agent

import (
	"context"
)

// Remote llm backend that serve model
type Provider interface {
	Chat(ctx context.Context, req CCReq) (*CCRes, error)
}
