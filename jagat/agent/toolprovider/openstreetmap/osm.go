package openstreetmap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
)

func init() {
	tooldef.Register("osm", NewTooldef)
}

type toolDefinition struct {
	def agent.Tool
}

func (d *toolDefinition) Def() agent.Tool {
	return d.def
}

var _ agent.XTool = (*OsmTool)(nil)

type OsmTool struct {
	cli *http.Client
}

type GeoCodeResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

func (osm *OsmTool) Call(ctx context.Context, fc agent.FunctionCall) (*agent.ToolResponse, error) {
	toolresponse := agent.ToolResponse{
		Name:   fc.Name,
		Output: map[string]any{},
	}

	if fc.Name != "open_street_map" {
		toolresponse.Output["error"] = "wrong function name, expected osm"
		return &toolresponse, nil
	}

	param := map[string]string{}
	if err := json.Unmarshal([]byte(fc.Arguments), &param); err != nil {
		toolresponse.Output["error"] = "wrong format arguemnts, expected {address : location_name}"
		return &toolresponse, nil
	}

	arg, ok := param["address"]
	if !ok {
		toolresponse.Output["error"] = "wrong format argument, address cannot be empty"
		return &toolresponse, nil
	}

	apiURL := "https://nominatim.openstreetmap.org/search?q=" + url.QueryEscape(arg) + "&format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Go Geocoding")
	resp, err := osm.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("osm failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result []GeoCodeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("osm failed marshal the response:%v", err)
	}

	if len(result) > 0 {
		// Get first result
		firstResult := result[0]
		toolresponse.Output["latitude"] = firstResult.Lat
		toolresponse.Output["longitude"] = firstResult.Lon
		return &toolresponse, nil
	}

	toolresponse.Output["error"] = fmt.Sprintf("no result found for the %s", fc.Arguments)
	return &toolresponse, nil

}

func (osm *OsmTool) Ping(ctx context.Context) error {
	return nil
}

// implement agent.ToolProvider
type toolProvider struct {
	agent.ToolDefinition
	agent.XTool
}

func NewTooldef(cft tooldef.Config) agent.ToolProvider {
	provider := &OsmTool{
		cli: http.DefaultClient,
	}

	definition := agent.Tool{
		Type: "Function",
		Function: agent.Function{
			Name:        "open_street_map",
			Description: "fetch geo location only (latitude, longitude), it will not give anything else",
			Parameters: agent.ParameterSchema{
				Type: agent.Parameter_Type_Object,
				Properties: map[string]agent.ParameterDefinition{
					"address": {
						Type:        "string",
						Description: "name or street or place. 'surabaya', '5th avenue, lenox' ",
					},
				},
				Required: []string{"address"},
			},
		},
	}

	tp := &toolProvider{
		ToolDefinition: &toolDefinition{
			def: definition,
		},
		XTool: provider,
	}

	return tp
}
