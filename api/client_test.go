package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/odit-bit/jagatai/api"
	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func basicRequest() *api.ChatRequest {
	return &api.ChatRequest{
		Content: []*api.Message{
			{
				Role: "user",
				Parts: []*agent.Part{
					api.NewTextPart("text"),
				},
			},
		},
	}
}

func basicResponse() *api.ChatResponse {
	return &api.ChatResponse{
		Text: "text-response",
	}
}

func Test_client(t *testing.T) {
	tTable := []struct {
		name         string
		requestFunc  func() *api.ChatRequest
		responseFunc func() *api.ChatResponse
		wantErr      string
	}{
		{
			name:         "success",
			requestFunc:  basicRequest,
			responseFunc: basicResponse,
		},
	}

	for _, tc := range tTable {
		t.Run(t.Name(), func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// read and optionally inspect the body
				var gotReq api.ChatRequest
				err := json.NewDecoder(r.Body).Decode(&gotReq)
				require.ErrorIs(t, nil, err)

				expectReq := tc.requestFunc()
				assert.Equal(t, expectReq, &gotReq)

				// ---- send mock response ----

				expectRes := tc.responseFunc()
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(expectRes); err != nil {
					t.Fatalf("failed to encode mock response: %v", err)
				}

			}))
			defer ts.Close()

			cli := api.NewClient(ts.URL, "test-key")
			ctx := context.Background()
			actRes, err := cli.Chat(ctx, *tc.requestFunc())
			require.ErrorIs(t, nil, err)
			assert.Equal(t, tc.responseFunc(), actRes)

		})
	}
}
