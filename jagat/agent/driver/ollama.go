package driver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/odit-bit/jagatai/jagat/agent"
	ollama "github.com/ollama/ollama/api"
)

//OpenAI compatible

const (
	_ollama_domain          = "http://127.0.0.1:11434"
	_ollama_completion_path = "v1/chat/completions"
)

//-----------------------------------------------

var _ agent.Provider = (*OllamaAPI)(nil)

type OllamaAPI struct {
	model string
	c     *ollama.Client
	conf  *Config
}

func NewOllamaAdapter(model string, key string, config *Config) (*OllamaAPI, error) {
	if model == "" {
		return nil, fmt.Errorf("ollama_adapter cannot be empty")
	}
	endpoint := config.Endpoint
	if endpoint == "" {
		endpoint = _ollama_domain
	}
	oUrl, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	cli := ollama.NewClient(oUrl, http.DefaultClient)
	oa := OllamaAPI{
		c:    cli,
		conf: config,
	}
	return &oa, err
}

// Chat implements LLM.
func (oapi *OllamaAPI) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	msgs := []ollama.Message{}
	for _, msg := range req.Messages {
		oMsg := ollama.Message{
			Role:    string(msg.Role),
			Content: msg.Text(),
		}

		msgs = append(msgs, oMsg)
	}

	// implement tools
	tools := []ollama.Tool{}
	for _, tool := range req.Tools {
		var t ollama.Tool
		OllamaTransformTool(tool, &t)
		tools = append(tools, t)
	}

	oReq := &ollama.ChatRequest{
		Model:    oapi.model,
		Messages: msgs,
		Stream:   &req.Stream,
		Think:    &req.Think,
		Options: map[string]any{
			"temperature": oapi.conf.Temperature,
			"top_p":       oapi.conf.TopP,
			"top_k":       oapi.conf.TopK,
			"min_p":       oapi.conf.MinP,
		},
		Tools: tools,
	}

	var resp *agent.CCRes
	err := oapi.c.Chat(ctx, oReq, func(cr ollama.ChatResponse) error {
		tcs := []*agent.ToolCall{}
		for _, tc := range cr.Message.ToolCalls {
			tcs = append(tcs, &agent.ToolCall{
				ID:   "",
				Type: tc.Function.Name,
				Function: agent.FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments.String(),
				},
			})
		}

		resp = &agent.CCRes{
			Model:   cr.Model,
			Created: cr.CreatedAt,
			Choices: []agent.Choice{
				{
					Text:         cr.Message.Content,
					FinishReason: cr.DoneReason,
					ToolCalls:    tcs,
				},
			},
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil

}

func (oapi *OllamaAPI) ChatGen(ctx context.Context) (any, error) {
	if err := oapi.c.Chat(ctx, &ollama.ChatRequest{}, func(cr ollama.ChatResponse) error {
		return nil
	}); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("not implemented")
}

// Transform takes a ToolA and produces the equivalent ToolB.
func OllamaTransformTool(aTool agent.Tool, bTool *ollama.Tool) {
	// var bTool ollama.Tool

	// 1) Copy the top‐level Type
	bTool.Type = aTool.Type

	// 2) Map the Function block
	bTool.Function.Name = aTool.Function.Name
	bTool.Function.Description = aTool.Function.Description

	// 3) Copy the Parameter “envelope” fields
	bTool.Function.Parameters.Type = aTool.Function.Parameters.Type
	bTool.Function.Parameters.Required = aTool.Function.Parameters.Required

	// 4) Initialize the B‐side properties map
	bTool.Function.Parameters.Properties = make(
		map[string]struct {
			Type        ollama.PropertyType `json:"type"`
			Items       any                 `json:"items,omitempty"`
			Description string              `json:"description"`
			Enum        []any               `json:"enum,omitempty"`
		},
	)

	// 5) Walk A’s Properties → build B’s Properties
	for propName, pa := range aTool.Function.Parameters.Properties {
		// convert []string → []any
		var enumAny []any
		for _, e := range pa.Enum {
			enumAny = append(enumAny, e)
		}

		// wrap the single‐string type into B.PropertyType (which is []string)
		pt := ollama.PropertyType{pa.Type}

		bTool.Function.Parameters.Properties[propName] = struct {
			Type        ollama.PropertyType `json:"type"`
			Items       any                 `json:"items,omitempty"`
			Description string              `json:"description"`
			Enum        []any               `json:"enum,omitempty"`
		}{
			Type:        pt,
			Description: pa.Description,
			Enum:        enumAny,
		}
	}

}
