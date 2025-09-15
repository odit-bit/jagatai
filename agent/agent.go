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
	tp       ToolsMap

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

// hold the tools
type ToolsMap map[string]Tool

func (t ToolsMap) Invoke(ctx context.Context, fc FunctionCall) (*ToolResponse, error) {
	tool, ok := t[fc.Name]
	if !ok {
		return nil, fmt.Errorf("tools not found")
	} else {
		if tool.call == nil {
			return nil, fmt.Errorf("tool failed invoke function: nil function")
		}
		res, err := tool.call(ctx, fc)
		if err != nil {
			return nil, err
		}
		return res, nil
	}
}

func (tm ToolsMap) ToSlice() []Tool {
	tools := []Tool{}
	for _, tool := range tm {
		tools = append(tools, tool)
	}
	return tools
}

type CompletionOptions struct {
	Think  bool
	Stream bool
}

func (agent *Agent) Completions(ctx context.Context, in CCReq) (*CCRes, error) {
	currentHistory := make([]Message, len(in.Messages))
	copy(currentHistory, in.Messages)

	for i := 0; i < agent.toolMaxCall; i++ {
		resp, err := agent.provider.Chat(ctx, CCReq{
			Model:      agent.model,
			Messages:   currentHistory,
			Tools:      agent.tp.ToSlice(),
			ToolChoice: "auto",
		})

		if err != nil {
			return nil, fmt.Errorf("agent provider: %v", err)
		}

		if !resp.IsToolCall() {
			return resp, nil
		}

		// TOOL CALL
		for _, tc := range resp.Choices[0].Message.Toolcalls {
			toolResp, err := agent.tp.Invoke(ctx, tc.Function)
			if err != nil {
				toolResp = &ToolResponse{map[string]any{
					"error": fmt.Sprintf("error: %s function failed to invoke", tc.Function.Name),
				}}
			}
			slog.Debug("agent_tool_call", "function", tc.Function, "error", err)

			toolRespMessage := Message{
				Role:         "tool",
				ToolCallID:   tc.ID,
				ToolResponse: toolResp,
			}
			currentHistory = append(currentHistory, toolRespMessage)
		}
	}

	return nil, fmt.Errorf("max tool calls exceed")
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

// Helper function to set provider on Agent for testing purposes.
func (a *Agent) SetProvider(p Provider) {
	a.provider = p
}
