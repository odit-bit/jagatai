package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/odit-bit/jagatai/jagat/agent"
	"google.golang.org/genai"
)

var _ agent.Provider = (*GeminiAdapter)(nil)

type GeminiAdapter struct {
	model string
	cli   *genai.Client
	conf  *Config
}

func NewGeminiAdapter(ctx context.Context, model, key string, config *Config) (*GeminiAdapter, error) {
	if model == "" {
		return nil, fmt.Errorf("gemini_adapter model cannot be empty")
	}

	cli, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  key,
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			ExtraBody: map[string]any{},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed start gemini_adapter: %s", err)
	}

	ga := &GeminiAdapter{
		model: model,
		cli:   cli,
		conf:  config,
	}

	return ga, nil
}

// Chat implements agent.Provider.
func (g *GeminiAdapter) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	var sys *genai.Content
	contents := []*genai.Content{}

	//message encoding to genai content
	for _, msg := range req.Messages {
		content := &genai.Content{}
		switch msg.Role {
		case agent.RoleAssistant:
			content.Role = genai.RoleModel

		case agent.RoleTool, agent.RoleUser:
			content.Role = genai.RoleUser

		case agent.RoleSystem:

		default:
			return nil, fmt.Errorf("gemini_adapter unknown message role: %v", msg.Role)

		}

		if err := messageToContent(msg, content); err != nil {
			return nil, fmt.Errorf("gemini_adapter failed convert message: %w", err)
		}

		if msg.Role == agent.RoleSystem {
			sys = content
			continue
		}

		contents = append(contents, content)
	}

	if len(contents) == 0 {
		return nil, fmt.Errorf("gemini_adapter content is empty")
	}

	config := genai.GenerateContentConfig{
		SystemInstruction: sys,
		Tools:             toolEncoding(req.Tools),
		SafetySettings:    safetySetting,
		Temperature:       g.conf.Temperature,
		TopP:              g.conf.TopP,
		TopK:              g.conf.TopK,
	}
	resp, err := g.cli.Models.GenerateContent(ctx, g.model, contents, &config)
	if err != nil {
		return nil, fmt.Errorf("genai_adapater failed generating content: %w", err)
	}

	// message decoding from genai.Content

	textPart := resp.Text()

	toolCall := []*agent.ToolCall{}
	if resp.FunctionCalls() != nil {
		for _, v := range resp.FunctionCalls() {
			tc, err := toToolCall(v)
			if err != nil {
				return nil, fmt.Errorf("gemini_adapter failed conversion function call: %v", err)
			}
			toolCall = append(toolCall, tc)
		}
	}

	//find text or toolCall
	// respons message
	candidate := resp.Candidates[0]
	a := &agent.CCRes{
		ID:    resp.ResponseID,
		Model: resp.ModelVersion,
		Choices: []agent.Choice{
			{
				Index:        0,
				Text:         textPart,
				ToolCalls:    toolCall,
				FinishReason: string(candidate.FinishReason),
			},
		},
		Created: resp.CreateTime,
		Usage: agent.Usage{
			CompletionTokens: candidate.TokenCount,
		},
	}

	return a, nil
}

func TestMessageToContent(src *agent.Message, dst *genai.Content) error {
	return messageToContent(src, dst)
}

func messageToContent(src *agent.Message, dst *genai.Content) error {

	for _, p := range src.Parts {
		part := &genai.Part{}
		var err error
		if p.Text != "" {
			part = genai.NewPartFromText(p.Text)
		} else if p.Blob != nil {
			part = genai.NewPartFromBytes(
				p.Blob.Bytes,
				p.Blob.Mime,
			)

		} else if p.Toolcall != nil {
			part.FunctionCall, err = fromToolCall(p.Toolcall)

		} else if p.ToolResponse != nil {
			part = genai.NewPartFromFunctionResponse(
				p.ToolResponse.Name,
				p.ToolResponse.Output,
			)
		}

		if err != nil {
			return err
		}

		dst.Parts = append(dst.Parts, part)
	}

	return nil
}

func toolEncoding(src []agent.Tool) []*genai.Tool {
	// Tool conversion
	tools := []*genai.Tool{}
	for _, rt := range src {
		tools = append(tools, &genai.Tool{
			FunctionDeclarations: []*genai.FunctionDeclaration{ToFunctionDeclaration(&rt)},
		})
	}
	return tools
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
func toToolCall(fc *genai.FunctionCall) (*agent.ToolCall, error) {
	if fc == nil {
		return nil, nil
	}

	// The genai.FunctionCall.Args is a map, but our ToolCall.Function.Arguments
	// is a string. We need to marshal the map into a JSON string.
	argumentsJSON, err := json.Marshal(fc.Args)
	if err != nil {
		return nil, fmt.Errorf("gemini_adapter failed marshal %T to JSON: %w", fc.Args, err)
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

func fromToolCall(tc *agent.ToolCall) (*genai.FunctionCall, error) {

	if tc == nil {
		return nil, nil
	}

	argsMap := map[string]any{}
	err := json.Unmarshal([]byte(tc.Function.Arguments), &argsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal agent tool call arguments to JSON: %v", err)
	}

	fc := genai.FunctionCall{
		ID:   tc.ID,
		Name: tc.Function.Name,
		Args: argsMap,
	}
	return &fc, nil
}

var safetySetting = []*genai.SafetySetting{
	{
		Category:  genai.HarmCategoryDangerousContent,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
	{
		Category:  genai.HarmCategoryHarassment,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
	{
		Category:  genai.HarmCategoryHateSpeech,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
	{
		Category:  genai.HarmCategorySexuallyExplicit,
		Threshold: genai.HarmBlockThresholdBlockNone,
	},
}
