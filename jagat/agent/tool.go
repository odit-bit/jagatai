package agent

import (
	"context"
	"fmt"
)

const (
	Parameter_Type_Object = "object"
)

// hold the tools
type Tools []ToolProvider

// type ToolProvider struct {
// 	Definition Tool
// 	Provider   XTool
// }

type ToolProvider interface {
	ToolDefinition
	XTool
}

type ToolDefinition interface {
	Def() Tool
}

func (t Tools) Invoke(ctx context.Context, fc FunctionCall) (*ToolResponse, error) {
	for _, tp := range t {
		def := tp.Def()
		if def.Function.Name == fc.Name {
			res, err := tp.Call(ctx, fc)
			if err != nil {
				return nil, err
			}
			return res, nil
		}
	}
	return nil, fmt.Errorf("tools not found")
}

func (tp Tools) Def() []Tool {
	copyDef := make([]Tool, len(tp))
	for i := range tp {
		copyDef[i] = tp[i].Def()
	}
	return copyDef
}

type XTool interface {
	// invoke the tool call
	Call(ctx context.Context, fn FunctionCall) (*ToolResponse, error)
	// the tool will not register and use if Ping method return error at build time.
	Ping(ctx context.Context) error
}

// Tool wraps a single tool entry.
// it will marshal into json before send to agent provider.
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function describes the function metadata and its parameter schema.
type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  ParameterSchema `json:"parameters"`
}

// ParameterSchema holds a minimal JSON‐schema for the function’s inputs.
type ParameterSchema struct {
	Type       string                         `json:"type"`
	Properties map[string]ParameterDefinition `json:"properties"`
	Required   []string                       `json:"required,omitempty"`
}

// ParameterDefinition defines each individual parameter in the schema.
type ParameterDefinition struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolCall represents one entry in the "tool_calls" array.
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// ToolResponse represent tool response entry in the message
type ToolResponse struct {
	//Tool name
	Name string
	// Tool response
	Output map[string]any
}

// FunctionCall holds the function name and its raw arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
