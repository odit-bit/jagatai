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
	provider Provider
	tp       ToolsMap

	toolMaxCall int
}

func New(provider Provider, opts ...OptionFunc) *Agent {

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

func (a *Agent) Completion(ctx context.Context, msgs []*Message) (*Message, error) {
	currentHistory := make([]*Message, len(msgs))
	copy(currentHistory, msgs)

	for i := 0; i < a.toolMaxCall; i++ {
		resp, err := a.provider.Chat(ctx, CCReq{
			// Model:      a.model,
			Messages:   currentHistory,
			Tools:      a.tp.ToSlice(),
			ToolChoice: "auto",
		})

		if err != nil {
			return nil, fmt.Errorf("agent provider: %v", err)
		}

		toolCalls, ok := resp.IsToolCall()
		if !ok {
			return NewTextMessage(RoleAssistant, resp.Choices[0].Text), nil
		}

		// TOOL CALL
		toolResMsg := Message{Role: RoleAssistant}
		for _, tc := range toolCalls {
			toolResp, err := a.tp.Invoke(ctx, tc.Function)
			if err != nil {
				toolResp = &ToolResponse{
					Name: tc.Function.Name,
					Output: map[string]any{
						"error": fmt.Sprintf("error: %s function failed to invoke", tc.Function.Name),
					}}
			}
			slog.Debug("agent_tool_call", "function", tc.Function, "error", err)

			toolResMsg.Parts = append(
				toolResMsg.Parts,
				&Part{
					ToolResponse: toolResp,
				},
			)

		}
		currentHistory = append(currentHistory, &toolResMsg)
	}

	return nil, fmt.Errorf("max tool calls exceed")
}
