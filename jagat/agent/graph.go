package agent

import (
	"context"
	"fmt"
	"log/slog"
)

// represent state that exchange between node.
type State struct {
	Message []*Message
}

// node is the unit of execution in the graph
type Node interface {
	// processing state and generate new state.
	Execute(ctx context.Context, state State) (next string, newState State, err error)
	// the name of the node, it will use by graph to determine which node is next.
	Name() string
}

// graph holds node and manage execution flow
type Graph struct {
	nodes map[string]Node
}

func NewGraph() *Graph {
	return &Graph{
		nodes: map[string]Node{},
	}
}

func (g *Graph) AddNode(node Node) {
	g.nodes[node.Name()] = node
}

// running execution
func (g *Graph) Run(ctx context.Context, entrypoint string, initState State) (*Message, error) {
	currentNode, ok := g.nodes[entrypoint]
	if !ok {
		return nil, fmt.Errorf("entrypoint node '%s' not found", entrypoint)
	}

	var nextNodeName string
	currentState := initState
	for {
		next, newState, err := currentNode.Execute(ctx, currentState)
		if err != nil {
			return nil, fmt.Errorf("failed executing node '%s' : %w", currentNode.Name(), err)
		}

		nextNodeName = next
		currentState = newState

		// the graph execution ends when it return empty string for next node
		if nextNodeName == "" || nextNodeName == "end" {
			finalMessage := currentState.Message[len(currentState.Message)-1]
			return finalMessage, nil
		}

		// next node
		nextNode, ok := g.nodes[nextNodeName]
		if !ok {
			return nil, fmt.Errorf("next node '%s' is not found", nextNodeName)
		}
		currentNode = nextNode
	}
}

// represent a single tool call.
type ToolNode struct {
	tp ToolProvider
}

func NewToolNode(tool ToolProvider) *ToolNode {
	return &ToolNode{
		tp: tool,
	}
}

func (tn *ToolNode) Name() string {
	return tn.tp.Def().Function.Name
}

func (tn *ToolNode) Execute(ctx context.Context, state State) (string, State, error) {
	lastMsg := state.Message[len(state.Message)-1]

	var toolResp *ToolResponse
	var err error

	tc, hasToolCall := lastMsg.ToolCall()
	if !hasToolCall {
		return "", state, fmt.Errorf("expected a tool call, but found none in the last message")
	}
	if tc.Function.Name != tn.Name() {
		return "", state, fmt.Errorf("routing error, expected tool call for '%s', but got '%s'", tn.Name(), tc.Function.Name)
	}

	toolResp, err = tn.tp.Call(ctx, tc.Function)
	if err != nil {
		toolResp = &ToolResponse{
			Name:   tn.Name(),
			Output: map[string]any{"error": err.Error()},
		}
	}
	slog.Debug("graph_nodes_tool", "tool", toolResp)

	if toolResp == nil {
		return "", state, fmt.Errorf("tool response is empty '%s'", tn.Name())
	}

	toolRespMsg := &Message{
		Role: RoleTool,
		Parts: []*Part{
			{ToolResponse: toolResp},
		},
	}
	state.Message = append(state.Message, toolRespMsg)

	return "agent", state, nil
}

type AgentNode struct {
	provider Provider
	tools    []Tool
}

func (an *AgentNode) Name() string {
	return "agent"
}

func (an *AgentNode) Execute(ctx context.Context, state State) (string, State, error) {
	resp, err := an.provider.Chat(ctx, CCReq{
		Messages: state.Message,
		Tools:    an.tools,
	})
	if err != nil {
		return "", state, err
	}

	modelMsg := Message{
		Role: RoleAssistant,
		Parts: []*Part{
			{Text: resp.Choices[0].Text},
		},
	}

	toolCalls, hasToolCall := resp.IsToolCall()
	if hasToolCall {
		// clear the text part for tool call
		modelMsg.Parts = []*Part{}
		for _, tc := range toolCalls {
			modelMsg.Parts = append(modelMsg.Parts, &Part{Toolcall: tc})
		}

	}

	state.Message = append(state.Message, &modelMsg)

	if hasToolCall {
		return toolCalls[0].Function.Name, state, nil
	}

	return "end", state, nil
}
