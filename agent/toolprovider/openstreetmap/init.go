package openstreetmap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/agent/tooldef"
)

func init() {
	tooldef.Register("osm", NewTooldef)
}

type GeoCodeResult struct {
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	DisplayName string `json:"display_name"`
}

var _ tooldef.Provider = (*OsmTool)(nil)

type OsmTool struct {
	cli *http.Client
}

func NewTooldef(cft tooldef.Config) tooldef.Provider {
	return &OsmTool{
		cli: http.DefaultClient,
	}
}

func (osm *OsmTool) Tooling() agent.Tool {
	t := agent.Tool{
		Type: "Function",
		Function: agent.Function{
			Name:        "open_street_map",
			Description: "fetch geo location info (latitude, longitude)",
			Parameters: agent.ParameterSchema{
				Type: agent.Parameter_Type_Object,
				Properties: map[string]agent.ParameterDefinition{
					"address": {
						Type:        "string",
						Description: "address of the place. example (jakarta or jakarta, indonesia)",
					},
				},
				Required: []string{"address"},
			},
		},
	}
	t.SetCallback(osm.callback)
	return t
}

func (osm *OsmTool) callback(ctx context.Context, fc agent.FunctionCall) (string, error) {
	if fc.Name != "open_street_map" {
		return "wrong function name, expected osm", nil
	}

	param := map[string]string{}
	if err := json.Unmarshal([]byte(fc.Arguments), &param); err != nil {
		return "wrong format arguemnts, expected {address : location_name}", nil
	}
	arg, ok := param["address"]
	if !ok {
		return "wrong format argument, address cannot be empty", nil
	}

	apiURL := "https://nominatim.openstreetmap.org/search?q=" + url.QueryEscape(arg) + "&format=json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Go Geocoding")
	resp, err := osm.cli.Do(req)
	if err != nil {
		return "", fmt.Errorf("osm failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	var result []GeoCodeResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("osm failed marshal the response:%v", err)
	}

	if len(result) > 0 {
		// Get first result
		firstResult := result[0]
		return fmt.Sprintf(
			"Latitude: %v Longitude: %v",
			firstResult.Lat,
			firstResult.Lon,
		), nil
	}

	return fmt.Sprintf("no result found for the %s", fc.Arguments), nil

}

func (osm *OsmTool) Ping(ctx context.Context) (bool, error) {
	return true, nil
}
