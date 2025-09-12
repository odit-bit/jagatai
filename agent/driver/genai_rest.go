package driver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/odit-bit/jagatai/agent"
	"google.golang.org/genai"
)

const (
	_genAI_domain           = "https://generativelanguage.googleapis.com"
	_genAI_completions_path = "v1beta/openai/chat/completions"
)

func init() {
	agent.RegisterDriver("genai_rest", NewGenaiAdapter)
}

// init simple OpenAI compatible api
func NewGenaiAdapter(key string) (agent.Provider, error) {
	e := endpoints{}
	e.Set(completionPath, _genAI_domain, _genAI_completions_path)

	return &Default{
		hc: http.DefaultClient,
		// domain:    _genAI_domain,
		apiKey:    key,
		maxRetry:  _http_default_max_retry,
		endpoints: e,
	}, nil
}

var _ agent.Provider = (*GenaiRestAdapter)(nil)

type GenaiRestAdapter struct {
	api *Default
}

// Chat implements agent.Provider.
func (g *GenaiRestAdapter) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {
	for _, v := range req.Messages {
		switch v.Role {
		case "assistant":
			v.Role = string(genai.RoleModel)
		case "system":
			v.Role = string(genai.RoleUser)
		case "tool":
			v.Role = string(genai.RoleUser)
		}
	}

	res, err := g.api.Chat(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("genai_rest_adapter failed generate content: %v", err)
	}
	for _, v := range res.Choices {
		v.Message.Role = "assistant"
	}
	return res, nil
}
