package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/odit-bit/jagatai/jagat/agent"
)

const (
	_openai_domain           = "https://api.openai.com"
	_openai_completions_path = "v1/chat/completions"
	_models_path             = "v1/models"
)

const (
	_http_default_max_retry = 2
)

const (
	completionPath canonical = "completion"
)

type canonical string

type endpoints map[canonical]string

func (e endpoints) Set(t canonical, domain, path string) {
	e[t] = fmt.Sprintf("%s/%s", domain, path)
}

func (e endpoints) Get(t canonical) string {
	return e[t]
}

var _ agent.Provider = (*Default)(nil)

// wrap simple OpenAI compatible api
type Default struct {
	model  string
	hc     *http.Client
	apiKey string
	// domain    string
	endpoints endpoints
	maxRetry  int

	config Config
}

func NewOpenAIAdapter( model, key string, config *Config) (agent.Provider, error) {
	if model == "" {
		return nil, fmt.Errorf("openai_adapter model cannot be empty")
	}

	e := endpoints{}
	e.Set(completionPath, _openai_domain, _openai_completions_path)
	return &Default{
		model: model,
		hc:    http.DefaultClient,
		// domain:    _openai_domain,
		apiKey:    key,
		maxRetry:  _http_default_max_retry,
		endpoints: e,
	}, nil
}

func (d *Default) doRety(req *http.Request) (*http.Response, error) {
	baseDelay := time.Duration(1 * time.Second)
	maxAttempt := d.maxRetry

	var errTry error
	for attempt := range maxAttempt {
		res, err := d.hc.Do(req)
		if err == nil && res.StatusCode < 400 {
			return res, nil
		}

		switch res.StatusCode {
		case 400, 404:
			return res, fmt.Errorf("driver failed send request status: %v", res.Status)
		}

		if res != nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}
		// exponential backoff with full jitter
		sleep := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		jitter := time.Duration(rand.Int63n(int64(sleep)))
		time.Sleep(jitter)
		slog.Debug("retry", "attempt", attempt, "max", maxAttempt, "status", res.StatusCode, "error", err)
		errTry = err
	}

	err1 := fmt.Errorf("api model request error: %v", errTry)
	err2 := fmt.Errorf("max attempt reach")
	errTry = errors.Join(err1, err2)

	return nil, errTry

}

// chat completions
func (d *Default) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, d.endpoints.Get(completionPath), bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	slog.Debug("provider request", "endpoint", request.URL.String())
	request.Header.Set("Content-Type", "application/json")
	// request.Header.Set("Accept", "application/x-ndjson")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

	resp, err := d.doRety(request)
	if err != nil {
		jsonData, _ := json.MarshalIndent(req, "", "  ")
		// if err != nil {
		// 	fmt.Println("Error marshalling to JSON:", err)
		// 	return nil, err
		// }
		// slog.Debug("openai_adapter", "request", string(jsonData))
		log.Println(string(jsonData))
		return nil, err
	}
	defer resp.Body.Close()

	//check error

	var ccr agent.CCRes
	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai_adapter failed read body: %v", err)
	}

	if err := json.Unmarshal(bb, &ccr); err != nil {
		return nil, fmt.Errorf("openai_adapter response error: %v", err)
	}

	if ccr.ID == "" {
		debugMap := map[string]any{}
		json.Unmarshal(bb, &debugMap)
		return nil, fmt.Errorf("openai_adapter response error: %v", debugMap)
	}

	return &ccr, nil

}
