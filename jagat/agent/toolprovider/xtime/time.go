package xtime

import (
	"context"
	"time"

	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
)

const (
	Namespace = "clock"
)

func init() {
	tooldef.Register(Namespace, NewTooldef)
}

var _ tooldef.Provider = (*clock)(nil)

type clock struct{}

func (o *clock) Tooling() agent.Tool {
	t := agent.Tool{
		Type: "function",
		Function: agent.Function{
			Name:        "get_current_time",
			Description: "get current time",
			Parameters: agent.ParameterSchema{
				Type: agent.Parameter_Type_Object,
				Properties: map[string]agent.ParameterDefinition{
					"Now": {
						Type:        "string",
						Description: "ask for current time occured with location set to UTC",
					},
				},
				Required: []string{},
			},
		},
	}

	t.SetCallback(func(ctx context.Context, fn agent.FunctionCall) (*agent.ToolResponse, error) {
		return &agent.ToolResponse{
			Output: map[string]any{
				"current_time_utc": time.Now().UTC().String(),
			},
		}, nil
	})

	return t
}

func (o *clock) Ping(_ context.Context) (bool, error) {
	return true, nil
}

func NewTooldef(cfg tooldef.Config) tooldef.Provider {
	return &clock{}
}
