package tavily

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_consturct_tavily(t *testing.T) {
	t.Setenv(ENV_API_KEY, "12345")
	var tav *Tavily

	tav, err := New(tooldef.Config{})
	require.NoError(t, err)

	assert.Equal(t, "12345", tav.key)
	t.Cleanup(func() {
		t.Setenv(ENV_API_KEY, "")
	})

	// ---

	cfg := tooldef.Config{
		Options: map[string]any{
			mapApikey: "12345",
		},
	}
	tav, err = New(cfg)
	require.NoError(t, err)
	assert.Equal(t, definition, tav.Def())
	assert.Equal(t, cfg.Options[mapApikey], tav.key)
}

func Test_request(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockResponse)
		require.NoError(t, err)
	}))
	defer ts.Close()

	tav, err := New(tooldef.Config{ApiKey: "12345", Endpoint: ts.URL})
	require.NoError(t, err)

	qr, err := tav.search(t.Context(), SearchParam{Query: "mock"})
	require.NoError(t, err)
	assert.Equal(t, mockResponse, *qr)
}

var mockResponse = QueryResponse{
	Query:  "today news in jakarta",
	Answer: "",
	Results: []SearchResult{
		{
			Title:      "The Jakarta Post - Still bold, fiercely independent",
			URL:        "https://www.thejakartapost.com/",
			Content:    "Indonesia · Archipelago. Riau Islands resumes live fish exports to Hong Kong with trial shipment ; Business · Economy. Govt declares local shrimp safe after joint",
			Score:      0.98598,
			RawContent: nil,
			Favicon:    "",
		},
		// {
		// 	Title:      "Jakarta Globe - breaking news today | Your City, Your World, Your ...",
		// 	URL:        "https://jakartaglobe.id/",
		// 	Content:    "Latest News · Indonesia Launches World's First Disaster Pooling Fund to Strengthen Crisis Financing · GoTo Gets $280 Million Loan Facility From DBS and UOB.",
		// 	Score:      0.98342,
		// 	RawContent: nil,
		// 	Favicon:    "",
		// },
		// {
		// 	Title:      "Jakarta - latest news, breaking stories and comment",
		// 	URL:        "https://www.the-independent.com/topic/jakarta",
		// 	Content:    "Jakarta · Indonesia Drug Arrests · News · <p>Plain‑clothed police officers are seen inside a villa where a shooting · Australasia",
		// 	Score:      0.9767,
		// 	RawContent: nil,
		// 	Favicon:    "",
		// },
		// {
		// 	Title:      "ANTARA News: Latest Indonesia News",
		// 	URL:        "https://en.antaranews.com/",
		// 	Content:    "Latest News · East Java Police detain nearly 1,000 over violent protests · Indonesia to speed up recognition of indigenous forests · Erick Thohir calls for united",
		// 	Score:      0.97373,
		// 	RawContent: nil,
		// 	Favicon:    "",
		// },
		// {
		// 	Title:      "Indonesia | Today's latest from Al Jazeera",
		// 	URL:        "https://www.aljazeera.com/where/indonesia/",
		// 	Content:    "Indonesian groups delay protests in Jakarta as police tighten security. Move comes as police set up checkpoints across Jakarta and deploy armoured vehicles to",
		// 	Score:      0.96968,
		// 	RawContent: nil,
		// 	Favicon:    "",
		// },
	},
}
