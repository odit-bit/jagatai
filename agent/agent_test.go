package agent_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/agent/tooldef"
	_ "github.com/odit-bit/jagatai/agent/toolprovider"
	"github.com/odit-bit/jagatai/agent/toolprovider/xtime"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ agent.Provider = (*mockProvider)(nil)

type mockProvider struct {
	ChatFunc func(ctx context.Context, req agent.CCReq) (*agent.CCRes, error)
}

func (mp *mockProvider) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {
	if mp.ChatFunc != nil {
		return mp.ChatFunc(ctx, req)
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("message cannot be nil")
	}
	query := req.Messages[len(req.Messages)-1]
	res := &agent.CCRes{
		Choices: []agent.Choice{
			{Message: query},
		},
	}

	if len(req.Tools) > 0 {
		if strings.Contains(query.Text, "time") {
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

// var req_with_tool = agent.CCReq{
// 	Messages: []agent.Message{
// 		{
// 			Text: "current time",
// 		},
// 	},
// }

func Test_agent_pipe(t *testing.T) {
	tp, err := tooldef.Build(t.Context(), []tooldef.Config{{Name: xtime.Namespace}})
	if err != nil {
		t.Fatal(err)
	}

	mp := mockProvider{}
	a, err := agent.NewPipe("test", &mp, agent.WithMaxToolCall(3), agent.WithTool(tp...))
	require.ErrorIs(t, err, nil)

	a.AddMiddleware(func(ctx context.Context, req *agent.CCReq, next agent.NextFunc) (*agent.CCRes, error) {
		// req.Messages = append(req.Messages, agent.Message{Role: "string", Content: "add by middleware"})
		if len(req.Messages) == 0 {
			return nil, fmt.Errorf("middleware: message cannot be nil")
		}
		return next(ctx, req)
	})

	req := agent.CCReq{
		Messages: []agent.Message{
			{
				Text: "current time",
			},
		},
	}

	res, err := a.Completions(t.Context(), req)
	if err != nil {
		t.Fatal(err)
	}
	_ = res
}

func TestAgent_Completions(t *testing.T) {
	tp, err := tooldef.Build(t.Context(), []tooldef.Config{{Name: xtime.Namespace}})
	require.NoError(t, err)

	testCases := []struct {
		name          string
		req           agent.CCReq
		provider      agent.Provider
		expectedText  string
		expectedError string
	}{
		{
			name: "successful completion with no tool call",
			req: agent.CCReq{
				Messages: []agent.Message{
					{Role: "user", Text: "hello"},
				},
			},
			provider: &mockProvider{
				ChatFunc: func(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {
					return &agent.CCRes{
						Choices: []agent.Choice{
							{Message: agent.Message{Text: "hello world"}},
						},
					}, nil
				},
			},
			expectedText: "hello world",
		},
		{
			name: "successful completion with one tool call",
			req: agent.CCReq{
				Messages: []agent.Message{
					{Role: "user", Text: "what time is it?"},
				},
			},
			provider: &mockProvider{
				ChatFunc: func(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {
					if len(req.Messages) == 1 { // First call
						return &agent.CCRes{
							Choices: []agent.Choice{
								{
									Message: agent.Message{
										Toolcalls: []agent.ToolCall{
											{
												ID:   "call_1",
												Type: "function",
												Function: agent.FunctionCall{
													Name:      "get_current_time",
													Arguments: `{"Now":"true"}`,
												},
											},
										},
									},
								},
							},
						}, nil
					}
					// Second call with tool response
					return &agent.CCRes{
						Choices: []agent.Choice{
							{Message: agent.Message{Text: "The time is now."}},
						},
					}, nil
				},
			},
			expectedText: "The time is now.",
		},
		{
			name: "error from provider",
			req: agent.CCReq{
				Messages: []agent.Message{
					{Role: "user", Text: "hello"},
				},
			},
			provider: &mockProvider{
				ChatFunc: func(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {
					return nil, errors.New("provider error")
				},
			},
			expectedError: "agent provider: provider error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// a, err := agent.NewWithProvider("test-model", "mp", "", agent.WithTool(tp...))
			// require.NoError(t, err)
			// a.SetProvider(tc.provider) // Helper function to set the provider for testing

			a := agent.New("test-model", tc.provider, agent.WithTool(tp...))

			res, err := a.Completions(context.Background(), tc.req.Messages)

			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Equal(t, tc.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedText, res.Choices[0].Message.Text)
			}
		})
	}
}
