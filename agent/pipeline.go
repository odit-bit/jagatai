package agent

import (
	"context"
)

type Payload struct {
	Request *CCReq
}

type NextFunc func(ctx context.Context, req *CCReq) (*CCRes, error)

type MiddlewareFunc func(ctx context.Context, req *CCReq, next NextFunc) (*CCRes, error)

type completionsPipeline struct {
	agent    *Agent
	pipeline []MiddlewareFunc
}

func NewPipe(model string, provider Provider, opts ...OptionFunc) (*completionsPipeline, error) {
	a := New(model, provider, opts...)

	p := completionsPipeline{
		agent:    a,
		pipeline: []MiddlewareFunc{},
	}
	return &p, nil
}

func (cp *completionsPipeline) AddMiddleware(middleware ...MiddlewareFunc) {
	cp.pipeline = append(cp.pipeline, middleware...)
}

func (cp *completionsPipeline) Completions(ctx context.Context, in CCReq) (*CCRes, error) {

	final := NextFunc(func(ctx context.Context, final *CCReq) (*CCRes, error) {
		return cp.agent.Completions(ctx, final.Messages)
	})

	chain := final
	for _, currentMiddleware := range cp.pipeline {
		next := chain
		chain = NextFunc(func(ctx context.Context, curIn *CCReq) (*CCRes, error) {
			return currentMiddleware(ctx, curIn, next)
		})
	}

	return chain(ctx, &in)
}
