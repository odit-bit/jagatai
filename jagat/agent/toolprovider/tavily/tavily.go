package tavily

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
)

func init() {
	tooldef.Register("tavily", New)
}

var definition = agent.Tool{
	Type: "function",
	Function: agent.Function{
		Name:        "web_search",
		Description: "search engine for retrieve data according to query",
		Parameters: agent.ParameterSchema{
			Type: agent.Parameter_Type_Object,
			Properties: map[string]agent.ParameterDefinition{
				"query": {
					Type:        "string",
					Description: "The search query. Must be a clear, concise question or topic.",
				},
				"search_depth": {
					Type:        "string",
					Description: "The depth of the search. Can be 'basic' or 'advanced'. Defaults to 'basic'.",
					Enum:        []string{"basic", "advanced"},
				},
				"max_results": {
					Type:        "integer",
					Description: "The maximum number of search results to return. Must be between 1 and 3.", // Provide constraints
				},
			},
			Required: []string{"query"},
		},
	},

	//
}

var _ agent.XTool = (*tavily)(nil)

type tavily struct {
	key  string
	tool agent.Tool
}

func New(cfg tooldef.Config) agent.ToolProvider {
	if cfg.ApiKey == "" {
		panic("tavily api key is empty")
	}
	t := tavily{
		key:  cfg.ApiKey,
		tool: definition,
	}
	return &t
}

func (t *tavily) Def() agent.Tool {
	return t.tool
}

type SearchParam struct {
	Query          string `json:"query"`
	AutoParameters bool   `json:"auto_parameters,omitempty"`
	Topic          string `json:"topic,omitempty"`
	// SearchDepth       int    `json:"search_depth,omitempty"`
	Chunks_per_source int  `json:"chunk_per_sources,omitempty"`
	MaxResults        int  `json:"max_results,omitempty"`
	Days              int  `json:"days,omitempty"`
	IncludeAnswer     bool `json:"include_answer,omitempty"`
	IncludeRawContent bool `json:"include_raw_content,omitempty"`
	IncludeImages     bool `json:"include_raw_images,omitempty"`
	Country           bool `json:"country,omitempty"`
}

func (sp *SearchParam) validate() error {
	if sp.Query == "" {
		return fmt.Errorf("search.query cannot be empty")
	}
	return nil
}

type SearchResult struct {
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Content    string  `json:"content"`
	Score      float64 `json:"score"`
	RawContent any     `json:"raw_content"` // can be nil or any type
	Favicon    string  `json:"favicon"`
}

type AutoParameters struct {
	Topic       string `json:"topic"`
	SearchDepth string `json:"search_depth"`
}

type QueryResponse struct {
	Query          string         `json:"query,omitempty"`
	Answer         string         `json:"answer,omitempty"`
	Images         []any          `json:"images,omitempty"`
	Results        []SearchResult `json:"results,omitempty"`
	AutoParameters AutoParameters `json:"auto_parameters,omitempty"`
	ResponseTime   string         `json:"response_time,omitempty"`
	RequestID      string         `json:"request_id,omitempty"`
}

// Call implements agent.XTool.
func (t *tavily) Call(ctx context.Context, fc agent.FunctionCall) (*agent.ToolResponse, error) {

	// send api request to tavilys

	var sp SearchParam
	if err := json.Unmarshal([]byte(fc.Arguments), &sp); err != nil {
		return nil, err
	}
	if err := sp.validate(); err != nil {
		return nil, err
	}

	// qp, err := t.search(ctx, sp)
	// if err != nil {
	// 	return nil, err
	// }
	qp := mockResponse

	return &agent.ToolResponse{Name: fc.Name, Output: map[string]any{
		"query":  qp.Query,
		"answer": qp.Results,
	}}, nil

}

func (t tavily) Ping(ctx context.Context) error {
	return nil
}

func (t *tavily) search(ctx context.Context, param SearchParam) (*QueryResponse, error) {
	url := "https://api.tavily.com/search"

	b, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", t.key)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode >= 399 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed calling '%s' tool, error: %v", t.tool.Function.Name, string(body))
	}
	var qr QueryResponse
	json.NewDecoder(res.Body).Decode(&qr)

	// if err := json.Unmarshal(body, &qr); err != nil {
	// 	return nil, err
	// }

	return &qr, nil

}
