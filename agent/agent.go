package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

const (
	toolMaxCallDefault = 3
)

type Agent struct {
	mx       sync.Mutex
	model    string
	provider Provider
	tp       ToolProviders

	toolMaxCall int
}

func New(model string, provider Provider, opts ...OptionFunc) *Agent {
	if model == "" {
		panic("model cannot be empty")
	}

	o := options{
		toolMaxCall: toolMaxCallDefault,
	}

	for _, fn := range opts {
		fn(&o)
	}

	if o.tools == nil {
		o.tools = ToolsMap{}
	}

	a := &Agent{
		mx:          sync.Mutex{},
		model:       model,
		provider:    provider,
		tp:          o.tools,
		toolMaxCall: o.toolMaxCall,
	}

	return a
}

type CompletionOptions struct {
	Think  bool
	Stream bool
}

type CompletionInput struct {
	Content string
}

func (agent *Agent) Completions(ctx context.Context, req *CCReq) (*CCRes, error) {
	req.Model = agent.model
	req.Tools = agent.tp.ToSlice()
	req.ToolChoice = "auto"

	resp, err := agent.provider.Chat(ctx, *req)
	if err != nil {
		return nil, fmt.Errorf("agent provider: %v", err)
	}

	// TOOL CALL
	for i := 0; i < agent.toolMaxCall && resp.IsToolCall(); i++ {
		for _, tc := range resp.Choices[0].Message.Toolcalls {
			toolResp, err := agent.tp.Invoke(ctx, tc)
			if err != nil {
				toolResp = fmt.Sprintf("error: %s function failed to invoke", tc.Function.Name)
			}
			slog.Debug("agent_tool_call", "function", tc.Function, "error", err)

			req.Messages = append(req.Messages, resp.Choices[0].Message, Message{
				Role:       "tool",
				Content:    toolResp,
				ToolCallID: tc.ID,
				Toolcalls:  []ToolCall{tc},
			})

		}

		// call the provider with tools response
		resp, err = agent.provider.Chat(ctx, *req)
		if err != nil {
			return nil, fmt.Errorf("agent provider (turn %d): %v", i+1, err)
		}
	}

	return resp, nil
}

func (agent *Agent) SetModel(model string) error {
	if ok := agent.mx.TryLock(); !ok {
		slog.Warn("agent failed to get lock")
		return fmt.Errorf("agent failed to get lock")
	}
	defer agent.mx.Unlock()
	agent.model = model
	return nil
}
