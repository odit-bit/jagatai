package driver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/odit-bit/jagatai/agent"
	ollama "github.com/ollama/ollama/api"
)

//OpenAI compatible

const (
	_ollama_domain          = "http://127.0.0.1:11434"
	_ollama_completion_path = "v1/chat/completions"
)

func init() {
	agent.RegisterDriver("ollama", NewOllamaAdapter)
}

// init simple OpenAI compatible api
func NewOllamaAdapter(key string) (agent.Provider, error) {
	e := endpoints{}
	e.Set(completionPath, _ollama_domain, _ollama_completion_path)

	return &Default{
		hc: http.DefaultClient,
		// domain:    _ollama_domain,
		apiKey:    key,
		maxRetry:  _http_default_max_retry,
		endpoints: e,
	}, nil
}

//-----------------------------------------------

var _ agent.Provider = (*OllamaAPI)(nil)

type OllamaAPI struct {
	// conf OllamaConfig
	c *ollama.Client
}

// Chat implements LLM.
func (oapi *OllamaAPI) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	msgs := []ollama.Message{}
	for _, msg := range req.Messages {
		msgs = append(msgs, ollama.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// implement tools
	tools := []ollama.Tool{}
	for _, tool := range req.Tools {
		var t ollama.Tool
		OllamaTransformTool(tool, &t)
		tools = append(tools, t)
	}

	oReq := &ollama.ChatRequest{
		Model:    req.Model,
		Messages: msgs,
		Stream:   &req.Stream,
		Think:    &req.Think,
		Options:  nil,
		Tools:    tools,
	}

	resp := new(agent.CCRes)
	err := oapi.c.Chat(ctx, oReq, func(cr ollama.ChatResponse) error {
		tcs := []agent.ToolCall{}
		for _, tc := range cr.Message.ToolCalls {
			tcs = append(tcs, agent.ToolCall{
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
			Created: agent.Timestamp(cr.CreatedAt.Unix()),
			Choices: []agent.Choice{
				{
					FinishReason: cr.DoneReason,
					Message: agent.Message{
						Role:      cr.Message.Role,
						Content:   cr.Message.Content,
						Toolcalls: tcs,
					}},
			},
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return resp, nil

}

func NewOllama(addr string) *OllamaAPI {

	e, _ := url.Parse(addr)
	c := ollama.NewClient(e, http.DefaultClient)

	// def := NewDefault(conf.OllamaEndpoint)
	oa := OllamaAPI{
		c: c,
	}

	return &oa
}

func (oapi *OllamaAPI) Models(ctx context.Context) (*agent.Models, error) {
	res, err := oapi.c.List(ctx)
	if err != nil {
		return nil, err
	}
	models := agent.Models{}
	for _, v := range res.Models {
		models.Data = append(models.Data, agent.Model{
			ID:      v.Name,
			Object:  "model",
			Created: agent.Timestamp(v.ModifiedAt.Unix()),
			OwnedBy: "library",
		})
	}
	return &models, nil
}

func (oapi *OllamaAPI) ChatGen(ctx context.Context) (any, error) {
	if err := oapi.c.Chat(ctx, &ollama.ChatRequest{}, func(cr ollama.ChatResponse) error {
		return nil
	}); err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("not implemented")
}

// return false if dst is not supported.
func TransformTool(tool agent.Tool, dst any) bool {
	switch v := dst.(type) {
	case *ollama.Tool:
		OllamaTransformTool(tool, v)
		return true
	default:
		return false
	}
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
