package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/odit-bit/jagatai/agent"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/option"
)

type GoogleSearch struct {
	svc  *customsearch.Service
	cxID string
}

// type QueryInput struct {
// 	Country    string
// 	Value      string
// 	Count      string
// 	SearchLang string `json:"search_lang"`
// }

// WIP DON'T USE !!
func NewGoogleSearch(ctx context.Context, apiKey, cxID string) (*GoogleSearch, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("failed init Google search: google api key is empty")
	}
	if cxID == "" {
		return nil, fmt.Errorf("failed init Google search: google search engine ID is empty")
	}
	cse, err := customsearch.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}

	gs := GoogleSearch{
		svc: cse,
	}
	return &gs, nil
}

type SearchResult struct {
	Title   string
	Snippet string
}

func (sr *SearchResult) String() string {
	return fmt.Sprintf("%s:%s\n%s:%s", "title", sr.Title, "snippet", sr.Snippet)
}

func (gs *GoogleSearch) SearchCallback(ctx context.Context, fc agent.FunctionCall) (string, error) {
	slog.Info("SearchCallback", "name", fc.Name, "args", fc.Arguments)
	q := map[string]string{}
	json.Unmarshal([]byte(fc.Arguments), &q)
	if q["query"] == "" {
		return "query not found", nil
	}

	// resp, err := gs.svc.Cse.List().Q(q["query"]).Do()
	resp, err := gs.svc.Cse.List().Cx(gs.cxID).Safe("off").Num(5).Q(q["query"]).Do()
	if err != nil {
		return "", err
	}

	builder := strings.Builder{}
	for i := 0; i < len(resp.Items); i++ {
		item := resp.Items[i]
		builder.WriteString(fmt.Sprintf("%s:%s\n%s:%s\n%s:%s", "Title", item.Title, "Snippet", item.Snippet, "Link", item.Link))
		builder.WriteString("\n")
		i++
	}
	return builder.String(), nil
}

func (gs *GoogleSearch) Tool() agent.Tool {
	t := agent.Tool{
		Type: "function",
		Function: agent.Function{
			Name:        "web_search",
			Description: "Get web related search results according to given query",
			Parameters: agent.ParameterSchema{
				Type: agent.Parameter_Type_Object,
				Properties: map[string]agent.ParameterDefinition{
					"query": {
						Type:        "string",
						Description: "query to search",
					},
				},
				Required: []string{"query"},
			},
		},
	}

	t.SetCallback(gs.SearchCallback)
	return t
}
