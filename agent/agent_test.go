package agent_test

import (
	"context"
	"strings"
	"testing"

	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/agent/tooldef"
	_ "github.com/odit-bit/jagatai/agent/toolprovider"
	"github.com/odit-bit/jagatai/agent/toolprovider/xtime"
)

var _ agent.Provider = (*mockProvider)(nil)

type mockProvider struct{}

func (mp *mockProvider) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {
	query := req.Messages[len(req.Messages)-1]
	res := &agent.CCRes{
		Choices: []agent.Choice{
			{Message: query},
		},
	}
	if len(req.Tools) > 0 {

		if strings.Contains(query.Content, "time") {
			tools := req.Tools
			res.Choices[0].Message = agent.Message{
				Toolcalls: []agent.ToolCall{
					{Function: agent.FunctionCall{
						Name:      tools[0].Function.Name,
						Arguments: "now",
					}},
				},
			}
		}
	}

	return res, nil

}

var req = agent.CCReq{
	Messages: []agent.Message{
		{
			Role:    "user",
			Content: "test-1",
		},
	},
}

func Test_agent(t *testing.T) {
	mp := mockProvider{}
	a := agent.New("test", &mp, agent.WithMaxToolCall(3))

	res, err := a.Completions(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res
	if res.Choices[0].Message.Content != req.Messages[0].Content {
		t.Fatalf("got result %s, expected %s", res.Choices[0].Message.Content, req.Messages[0].Content)
	}
}

var req_tool = agent.CCReq{
	Messages: []agent.Message{
		{
			Role:    "user",
			Content: "current time",
		},
	},
}

func Test_agent_tools(t *testing.T) {
	tp, err := tooldef.Build(t.Context(), []tooldef.Config{{Name: xtime.Namespace}})
	if err != nil {
		t.Fatal(err)
	}
	_ = tp
	a := agent.New("test", &mockProvider{}, agent.WithTool(tp...))

	res, err := a.Completions(t.Context(), req_tool)
	if err != nil {
		t.Fatal(err)
	}
	_ = res

}
