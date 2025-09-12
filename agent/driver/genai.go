package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/odit-bit/jagatai/agent"
	"google.golang.org/genai"
)

func init() {
	agent.RegisterDriver("genai", NewGeminiAdapter)
}

var _ agent.Provider = (*GeminiAdapter)(nil)

type GeminiAdapter struct {
	cli *genai.Client
}

func NewGeminiAdapter(key string) (agent.Provider, error) {
	cli, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed start gemini_adapter: %s", err)
	}
	return &GeminiAdapter{cli: cli}, nil
}

// Chat implements agent.Provider.
func (g *GeminiAdapter) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	// jsonRequest, _ := json.MarshalIndent(req, "", "  ")
	// log.Println("request: ", string(jsonRequest))

	content := []*genai.Content{}

	sys := genai.Content{}
	// Message conversion
	for _, c := range req.Messages {
		parts := []*genai.Part{}
		var role string

		switch c.Role {
		case "system":
			sys.Parts = []*genai.Part{{Text: c.Content}}
			continue // Skip appending to content
		case "assistant":
			role = genai.RoleModel
			if c.Content != "" {
				part := genai.NewPartFromText(c.Content)
				parts = append(parts, part)
			} else if len(c.Toolcalls) != 0 {
				part2 := genai.NewPartFromFunctionCall(
					c.Toolcalls[0].Function.Name,
					map[string]any{"arguments": c.Toolcalls[0].Function.Arguments},
				)
				parts = append(parts, part2)
			} else {
				continue // Skip empty assistant messages that aren't tool calls
			}

		case "tool":
			part := genai.NewPartFromFunctionResponse(
				c.Toolcalls[0].Function.Name,
				map[string]any{"output": c.Content},
			)
			parts = append(parts, part)
		case "user":
			role = genai.RoleUser
			part := genai.NewPartFromText(c.Content)
			parts = append(parts, part)
		}

		content = append(content, &genai.Content{
			Parts: parts,
			Role:  role,
		})
	}

	// jsonResult, _ := json.MarshalIndent(content, "", "  ")
	// log.Println("result: ", string(jsonResult))

	// Tool conversion
	tools := []*genai.Tool{}
	for _, rt := range req.Tools {
		tools = append(tools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{ToFunctionDeclaration(&rt)},
		})
	}

	// execution call
	resp, err := g.cli.Models.GenerateContent(ctx, req.Model, content, &genai.GenerateContentConfig{
		Tools:             tools,
		SystemInstruction: &sys,
		ThinkingConfig: &genai.ThinkingConfig{
			IncludeThoughts: req.Think,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("gemini_adapter failed generate content: %v", err)
	}

	//conversion back the message and the tools
	//toolcall
	toolCall := []agent.ToolCall{}
	text := ""
	if resp.FunctionCalls() != nil {
		for _, v := range resp.FunctionCalls() {
			tc, err := ToToolCall(v)
			if err != nil {
				return nil, fmt.Errorf("gemini_adapter failed conversion function call: %v", err)
			}
			toolCall = append(toolCall, *tc)
		}
	} else {
		text = resp.Text()
	}

	// respons message
	a := &agent.CCRes{
		ID:    resp.ResponseID,
		Model: resp.ModelVersion,
		Choices: []agent.Choice{
			{
				Message: agent.Message{
					Role:      "assistant",
					Content:   text,
					Toolcalls: toolCall,
				},
			},
		},
		Created: resp.CreateTime,
	}

	// jsonResult, _ := json.MarshalIndent(a, "", "  ")
	// log.Println("result: ", string(jsonResult))

	return a, nil
}

// ToFunctionDeclaration converts a Tool into a genai.FunctionDeclaration.
// This version correctly maps data types to the format required by the Google API.
func ToFunctionDeclaration(t *agent.Tool) *genai.FunctionDeclaration {
	if t == nil {
		return nil
	}

	// Helper function to map type strings to the API's required types.
	mapType := func(inputType string) genai.Type {
		switch strings.ToLower(inputType) {
		case "string":
			return genai.TypeString
		case "number", "float", "double": // Handles float, double, etc.
			return genai.TypeNumber
		case "integer", "int":
			return genai.TypeInteger
		case "boolean", "bool":
			return genai.TypeBoolean
		case "object":
			return genai.TypeObject
		case "array":
			return genai.TypeArray
		default:
			// Fallback for any other types, converting to uppercase
			return genai.Type(strings.ToUpper(inputType))
		}
	}

	// Create the main parameter schema for the function declaration.
	paramSchema := &genai.Schema{
		Type:       mapType(t.Function.Parameters.Type),
		Properties: make(map[string]*genai.Schema),
		Required:   t.Function.Parameters.Required,
	}

	// Iterate over the properties and convert each one, using mapping logic.
	for name, propDef := range t.Function.Parameters.Properties {
		paramSchema.Properties[name] = &genai.Schema{
			Type:        mapType(propDef.Type),
			Description: propDef.Description,
			Enum:        propDef.Enum,
		}
	}

	// Construct the final FunctionDeclaration.
	declaration := &genai.FunctionDeclaration{
		Name:        t.Function.Name,
		Description: t.Function.Description,
		Parameters:  paramSchema,
	}
	return declaration
}

// ToToolCall converts a genai.FunctionCall into a ToolCall.
// It returns an error if the arguments map cannot be marshaled to a JSON string.
func ToToolCall(fc *genai.FunctionCall) (*agent.ToolCall, error) {
	if fc == nil {
		return nil, nil
	}

	// The genai.FunctionCall.Args is a map, but our ToolCall.Function.Arguments
	// is a string. We need to marshal the map into a JSON string.
	argumentsJSON, err := json.Marshal(fc.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal function call arguments to JSON: %w", err)
	}

	toolCall := &agent.ToolCall{
		ID:   fc.ID,
		Type: "function", // This is static since the source is a FunctionCall.
		Function: agent.FunctionCall{
			Name:      fc.Name,
			Arguments: string(argumentsJSON),
		},
	}

	return toolCall, nil
}
