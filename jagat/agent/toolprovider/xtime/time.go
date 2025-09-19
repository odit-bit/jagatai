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
	tooldef.Register(Namespace, NewToolProvider)
}

var _ agent.ToolProvider = (*clock)(nil)

type clock struct {
	def agent.Tool
}

func (c *clock) Ping(ctx context.Context) error {
	return nil
}

func (c *clock) Def() agent.Tool {
	return c.def
}

func NewToolProvider(cfg tooldef.Config) agent.ToolProvider {
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

	return &clock{
		def: t,
	}
}

func (o *clock) Call(ctx context.Context, fc agent.FunctionCall) (*agent.ToolResponse, error) {
	tr := &agent.ToolResponse{
		Output: map[string]any{
			"current_time_utc": time.Now().UTC().String(),
		},
	}

	return tr, nil
}
