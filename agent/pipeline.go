package agent

import (
	"context"
)

type NextFunc func(ctx context.Context, req *CCReq) (*CCRes, error)

type MiddlewareFunc func(ctx context.Context, req *CCReq, next NextFunc) (*CCRes, error)

type completionsPipeline struct {
	agent    *Agent
	pipeline []MiddlewareFunc
}

func NewPipe(model string, driver string, key string, opts ...OptionFunc) (*completionsPipeline, error) {
	a, err := NewWithProvider(model, driver, key, opts...)
	if err != nil {
		return nil, err
	}

	p := completionsPipeline{
		agent:    a,
		pipeline: []MiddlewareFunc{},
	}
	return &p, nil
}

func (cp *completionsPipeline) AddMiddleware(middleware ...MiddlewareFunc) {
	cp.pipeline = append(cp.pipeline, middleware...)
}

func (cp *completionsPipeline) Completions(ctx context.Context, req *CCReq) (*CCRes, error) {

	final := NextFunc(func(ctx context.Context, req *CCReq) (*CCRes, error) {
		return cp.agent.Completions(ctx, req)
	})

	chain := final
	for _, currentMiddleware := range cp.pipeline {
		next := chain
		chain = NextFunc(func(ctx context.Context, req *CCReq) (*CCRes, error) {
			return currentMiddleware(ctx, req, next)
		})
	}

	return chain(ctx, req)
}
