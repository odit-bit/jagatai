package openmeteo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/agent/tooldef"
)

var _ tooldef.Provider = (*WeatherTool)(nil)

type WeatherTool struct {
	// ctx    context.Context
	client   *http.Client
	endpoint string
}

// func NewMeteoWeather(ctx context.Context) *WeatherTool {
// 	urlEndpoint, _ := url.Parse("https://api.open-meteo.com/v1/forecast")
// 	wt := WeatherTool{
// 		// ctx:    ctx,
// 		client: http.DefaultClient,
// 		url:    urlEndpoint,
// 	}
// 	return &wt
// }

func NewTooldef(cfg tooldef.Config) tooldef.Provider {
	urlEndpoint, _ := strings.CutSuffix(cfg.Endpoint, "/")
	wt := WeatherTool{
		// ctx:    ctx,
		client:   http.DefaultClient,
		endpoint: urlEndpoint,
	}
	return &wt
}

func (wt *WeatherTool) Ping(ctx context.Context) (bool, error) {
	// endpoint := fmt.Sprintf("%s/v1/forecast", wt.endpoint)
	// resp, err := wt.client.Get(endpoint)
	// if err != nil {
	// 	return false, err
	// }
	// if resp.StatusCode > 299 {
	// 	return false, nil
	// }
	// defer resp.Body.Close()
	return true, nil

}

type MeteoResult struct {
	Timezone string
	Current  Current `json:"current"`
}

type Current struct {
	Time     Timestamp
	Interval int
	Temp2m   float64 `json:"temperature_2m"`
}

type Timestamp int64

func (t *Timestamp) Time() time.Time {
	return time.Unix(int64(*t), 0)
}

func (t *Timestamp) String() string {
	return t.Time().String()
}

func (wt *WeatherTool) do(req *http.Request) (*MeteoResult, error) {
	resp, err := wt.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather tool failed do request: %v", err)
	}
	if resp.StatusCode == 200 {
		defer resp.Body.Close()
		var mr MeteoResult
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(b, &mr); err != nil {
			return nil, err
		}

		return &mr, nil
	}
	return nil, fmt.Errorf("weather tool response status code: %v", resp.StatusCode)
}

func (wt *WeatherTool) GetCurrentWeather(ctx context.Context, lat, long float64) (*Current, error) {

	endpoint, _ := url.Parse(fmt.Sprintf("%s/v1/forecast", wt.endpoint))
	query := endpoint.Query()
	query.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	query.Set("longitude", strconv.FormatFloat(long, 'f', -1, 64))
	query.Set("current", "temperature_2m")
	query.Set("timeformat", "unixtime")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	mr, err := wt.do(req)
	if err != nil {
		return nil, err
	}
	return &mr.Current, nil
}

func (wt *WeatherTool) Callback(ctx context.Context, fc agent.FunctionCall) (string, error) {
	param := struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}{}
	if err := json.Unmarshal([]byte(fc.Arguments), &param); err != nil {
		return "", err
	}
	if param.Latitude == 0 || param.Longitude == 0 {
		return "", fmt.Errorf("latitude or longitude cannot be empty")
	}

	curr, err := wt.GetCurrentWeather(ctx, param.Latitude, param.Longitude)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(curr.Temp2m, 'f', -1, 64), nil
}

func (wt *WeatherTool) Tooling() agent.Tool {
	t := agent.Tool{
		Type: "function",
		Function: agent.Function{
			Name:        "get_current_weather",
			Description: "get current temperature (celcius) for provided coordinates. example {latitude: :-6.9218457, longitude:107.6070833}",
			Parameters: agent.ParameterSchema{
				Type: agent.Parameter_Type_Object,
				Properties: map[string]agent.ParameterDefinition{
					"latitude": {
						Type: "float",
					},
					"longitude": {
						Type: "float",
					},
				},
				Required: []string{"latitude", "longitude"},
			},
		},
	}

	t.SetCallback(wt.Callback)
	return t
}
