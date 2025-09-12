package agent

import (
	"context"
	"fmt"
)

const (
	Parameter_Type_Object = "object"
)

// Tool wraps a single tool entry.
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`

	call FunctionCallbackFunc
}

// TODO: it will introduce ambiguity as set callback function defined with method or when struct constructed
func (t *Tool) SetCallback(fn FunctionCallbackFunc) {
	t.call = fn
}

type FunctionCallbackFunc func(ctx context.Context, fn FunctionCall) (string, error)

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

// FunctionCall holds the function name and its raw arguments.
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolProviders interface {
	Invoke(ctx context.Context, tc ToolCall) (string, error)
	ToSlice() []Tool
}

///

type ToolsMap map[string]Tool

func (t ToolsMap) Invoke(ctx context.Context, tc ToolCall) (string, error) {
	tool, ok := t[tc.Function.Name]
	if !ok {
		return "", fmt.Errorf("tools not found")
	} else {
		if tool.call == nil {
			return "", fmt.Errorf("tool failed invoke function: nil function")
		}
		res, err := tool.call(ctx, tc.Function)
		if err != nil {
			return "", err
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

// func (t ToolsMap) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(t.slice())
// }

// func (t ToolsMap) UnmarshalJSON(b []byte) error {
// 	if t == nil {
// 		t = ToolsMap{}
// 	}
// 	return json.Unmarshal(b, &t)
// }
