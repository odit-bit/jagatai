package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/odit-bit/jagatai/agent"
	"google.golang.org/genai"
)

var _ agent.Provider = (*GeminiAdapter)(nil)

type GeminiAdapter struct {
	cli  *genai.Client
	conf *Config
}

func NewGeminiAdapter(ctx context.Context, key string, config *Config) (*GeminiAdapter, error) {
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
		cli:  cli,
		conf: config,
	}

	return ga, nil
}

// Chat implements agent.Provider.
func (g *GeminiAdapter) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	// jsonRequest, _ := json.MarshalIndent(req, "", "  ")
	// log.Println("request: ", string(jsonRequest))

	content := []*genai.Content{}
	sys := genai.Content{}

	// Message conversion to content
	for _, msg := range req.Messages {
		var c *genai.Content
		var err error

		switch msg.Role {
		case "system":
			sys.Parts = []*genai.Part{
				genai.NewPartFromText(msg.Text),
			}
			continue

		case "assistant":
			c, err = convertAssistant(&msg)

		case "user", "tool":
			c, err = convertUser(&msg)

		default:
			err = fmt.Errorf("unknown message type/role: %v", msg.Role)
		}

		if err != nil {
			return nil, fmt.Errorf("genai_adapter failed convert agent message: %v", err)
		}

		content = append(content, c)
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
		SafetySettings:    safetySetting,
		ResponseModalities: []string{
			string(genai.ModalityText),
		},
		TopP:        g.conf.TopP,
		TopK:        g.conf.TopK,
		Temperature: g.conf.Temperature,
	})

	if err != nil {
		return nil, fmt.Errorf("gemini_adapter failed generate content: %v", err)
	}

	//conversion back the message and the tools
	//toolcall
	toolCall := []agent.ToolCall{}
	if resp.FunctionCalls() != nil {
		for _, v := range resp.FunctionCalls() {
			tc, err := toToolCall(v)
			if err != nil {
				return nil, fmt.Errorf("gemini_adapter failed conversion function call: %v", err)
			}
			toolCall = append(toolCall, *tc)
		}
	}

	data := agent.Blob{
		Bytes: []byte{},
	}
	for _, v := range resp.Candidates[0].Content.Parts {
		if v.InlineData != nil {
			if v.InlineData.Data != nil {
				data.Bytes = v.InlineData.Data
				data.Mime = v.InlineData.MIMEType
			}
		}
	}

	// respons message
	a := &agent.CCRes{
		ID:    resp.ResponseID,
		Model: resp.ModelVersion,
		Choices: []agent.Choice{
			{
				Message: agent.Message{
					Role:      "assistant",
					Text:      resp.Text(),
					Toolcalls: toolCall,
					Data:      &data,
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
func toToolCall(fc *genai.FunctionCall) (*agent.ToolCall, error) {
	if fc == nil {
		return nil, nil
	}

	// The genai.FunctionCall.Args is a map, but our ToolCall.Function.Arguments
	// is a string. We need to marshal the map into a JSON string.
	argumentsJSON, err := json.Marshal(fc.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal genai function call arguments to JSON: %w", err)
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

// helper for convert agent assistant response message
func convertAssistant(msg *agent.Message) (*genai.Content, error) {
	parts := []*genai.Part{}

	//tool call
	for _, tc := range msg.Toolcalls {
		fc, err := fromToolCall(&tc)
		if err != nil {
			return nil, err
		}
		parts = append(parts, genai.NewPartFromFunctionCall(fc.Name, fc.Args))
	}

	if msg.Text != "" {
		parts = append(parts, genai.NewPartFromText(msg.Text))
	}

	c := genai.Content{
		Role:  genai.RoleModel,
		Parts: parts,
	}
	return &c, nil
}

func convertUser(msg *agent.Message) (*genai.Content, error) {

	Parts := []*genai.Part{}
	if msg.Text != "" {
		Parts = append(Parts, genai.NewPartFromText(msg.Text))
	}

	if msg.Data != nil {
		Parts = append(Parts, genai.NewPartFromBytes(
			msg.Data.Bytes,
			msg.Data.Mime,
		))
	}

	//add
	if msg.Toolcalls != nil {
		fcPart := genai.NewPartFromFunctionResponse(
			msg.Toolcalls[0].Function.Name,
			msg.ToolResponse.Output,
		)
		Parts = append(Parts, fcPart)

	}
	content := &genai.Content{
		Role:  genai.RoleUser,
		Parts: Parts,
	}
	return content, nil
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
