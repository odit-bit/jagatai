package jagat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAgent provides a mock implementation of the Agent interface for testing.
type mockAgent struct {
	CompletionsFunc func(ctx context.Context, msgs []*agent.Message) (*agent.Message, error)
}

// Completions implements the Agent interface for the mockAgent.
func (m *mockAgent) Completion(ctx context.Context, msgs []*agent.Message) (*agent.Message, error) {
	if m.CompletionsFunc != nil {
		return m.CompletionsFunc(ctx, msgs)
	}
	return agent.NewTextMessage("assistant", "mock response"), nil
}

func TestHandleAgentCompletions(t *testing.T) {
	// Setup
	e := echo.New()
	mockAgent := &mockAgent{}
	// Register the routes with the mock agent
	RestHandler(context.Background(), mockAgent, e)

	testCases := []struct {
		name               string
		requestBody        string
		contentType        string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name: "successful completion",
			requestBody: `{
				"messages": [{"role": "user", "text": "hello"}]
			}`,
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "mock response",
		},
		{
			name:               "bad request - invalid json",
			requestBody:        `{"messages": [`,
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad json format",
		},
		{
			name:               "bad request - wrong content type",
			requestBody:        `{}`,
			contentType:        echo.MIMETextPlain,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "expecting json body",
		},
		{
			name:               "bad request - empty body",
			requestBody:        "",
			contentType:        echo.MIMEApplicationJSON,
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   "bad json format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request and recorder
			req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(tc.requestBody))
			req.Header.Set(echo.HeaderContentType, tc.contentType)
			rec := httptest.NewRecorder()

			// Serve the HTTP request to the Echo instance
			e.ServeHTTP(rec, req)

			// Assertions
			require.Equal(t, tc.expectedStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tc.expectedResponse)
		})
	}
}
