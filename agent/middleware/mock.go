package middleware

import (
	"context"
	"log"

	"github.com/odit-bit/jagatai/agent"
)

func init() {
	Register("mock", NewMockMiddleware)
}

type MockMiddleware struct{}

func (mm *MockMiddleware) Run(ctx context.Context, req *agent.CCReq, next agent.NextFunc) (*agent.CCRes, error) {
	log.Println("Mock Middleware passthrough")
	return next(ctx, req)
}

func NewMockMiddleware(name string, config Config) (agent.MiddlewareFunc, error) {
	mm := MockMiddleware{}
	return mm.Run, nil
}
