package agent

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// orchestrate execution flow.
type Agent struct {
	mx       sync.Mutex
	provider Provider
	tools    Tools

	toolMaxCall int
}

func New(provider Provider, opts ...OptionFunc) *Agent {

	o := options{}

	for _, fn := range opts {
		fn(&o)
	}

	if o.tools == nil {
		o.tools = Tools{}
	}

	a := &Agent{
		mx:          sync.Mutex{},
		provider:    provider,
		tools:       o.tools,
		toolMaxCall: o.toolMaxCall,
	}

	return a
}

type CompletionOptions struct {
	Think  bool
	Stream bool
}

func (a *Agent) completionDag(ctx context.Context, msgs []*Message) (*Message, error) {

	graph := NewGraph()
	agentNode := AgentNode{
		provider: a.provider,
		tools:    a.tools.Def(),
	}
	graph.AddNode(&agentNode)

	for _, tool := range a.tools {
		toolNode := NewToolNode(tool)
		graph.AddNode(toolNode)
	}

	copyMsg := make([]*Message, len(msgs))
	copy(copyMsg, msgs)
	initState := State{
		Message: copyMsg,
	}

	return graph.Run(ctx, "agent", initState)
}

func (a *Agent) Completion(ctx context.Context, msgs []*Message) (*Message, error) {
	return a.completionDag(ctx, msgs)
}


//Deprecate, subjet to remove.
func (a *Agent) completion(ctx context.Context, msgs []*Message) (*Message, error) {
	currentHistory := make([]*Message, len(msgs))
	copy(currentHistory, msgs)

	for i := 0; i < a.toolMaxCall; i++ {
		resp, err := a.provider.Chat(ctx, CCReq{
			// Model:      a.model,
			Messages:   currentHistory,
			Tools:      a.tools.Def(),
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
			toolResp, err := a.tools.Invoke(ctx, tc.Function)
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
