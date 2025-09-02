package driver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/odit-bit/jagatai/agent"
)

const (
	_cc_path     = "v1/chat/completions"
	_models_path = "v1/models"

	_default_max_retry = 2
)

var _ agent.Provider = (*Default)(nil)

type Default struct {
	hc       *http.Client
	apiKey   string
	domain   string
	maxRetry int
}

// init simple OpenAI compatible api
func NewDefault(addr, key string) *Default {
	return &Default{
		hc:       http.DefaultClient,
		domain:   addr,
		apiKey:   key,
		maxRetry: _default_max_retry,
	}
}

func (d *Default) doRety(req *http.Request) (*http.Response, error) {
	baseDelay := time.Duration(1 * time.Second)
	maxAttempt := 3

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/x-ndjson")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

	var errTry error
	for attempt := range maxAttempt {
		res, err := d.hc.Do(req)
		if err == nil && res.StatusCode < 400 {
			return res, nil
		}
		if res != nil {
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}
		// exponential backoff with full jitter
		sleep := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
		jitter := time.Duration(rand.Int63n(int64(sleep)))
		time.Sleep(jitter)
		slog.Debug("retry", "attempt", attempt, "max", maxAttempt, "error", err)
		errTry = err
	}

	err1 := fmt.Errorf("api model request error: %v", errTry)
	err2 := fmt.Errorf("max attempt reach")
	errTry = errors.Join(err1, err2)

	return nil, errTry

}

// func (d *Default) ChatStream(ctx context.Context, req agent.CCReq) (<-chan agent.Message, error) {
// 	e := fmt.Sprintf("%s/%s", d.domain, _cc_path)

// 	b, err := json.Marshal(req)
// 	if err != nil {
// 		return nil, err
// 	}

// 	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e, bytes.NewBuffer(b))
// 	if err != nil {
// 		return nil, err
// 	}

// 	resp, err := d.doRety(request)
// 	if err != nil {
// 		return nil, err
// 	}

// 	msgC := make(chan agent.Message, 1)
// 	go func() {
// 		defer resp.Body.Close()
// 		defer close(msgC)

// 		scanner := bufio.NewScanner(resp.Body)
// 		var ccr agent.CCRes

// 		prefix := []byte("data: ") // accomodate ollama stream response
// 		_ = prefix

// 		scanBuffer := make([]byte, 512*humanize.KByte)
// 		scanner.Buffer(scanBuffer, 512*humanize.KByte)

// 		for scanner.Scan() {
// 			bts := scanner.Bytes()
// 			bts, _ = bytes.CutPrefix(bts, prefix)

// 			// BUG, always return syntaxError even though it unmarshaled
// 			if err := json.Unmarshal(bts, &ccr); err != nil {
// 				switch err.(type) {
// 				case *json.SyntaxError:
// 					continue

// 				default:
// 					return //fmt.Errorf("unmarshal error: %v type:%T", err, err)
// 				}

// 			}
// 			// return &ccr, fmt.Errorf("stream response err: %v", err)
// 			select {
// 			case <-ctx.Done():
// 				return
// 			case msgC <- ccr.Choices[0].Message:

// 			}
// 		}

// 	}()
// 	return msgC, nil
// }

// chat completions
func (d *Default) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

	e := fmt.Sprintf("%s/%s", d.domain, _cc_path)

	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, e, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	resp, err := d.doRety(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	//check error

	var ccr agent.CCRes
	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("driver failed read body: %v", err)
	}

	if err := json.Unmarshal(bb, &ccr); err != nil {
		return nil, fmt.Errorf("driver response error: %v", err)
	}

	if ccr.ID == "" {
		debugMap := map[string]any{}
		json.Unmarshal(bb, &debugMap)
		return nil, fmt.Errorf("driver response error: %v", debugMap)
	}

	return &ccr, nil

}

func (d *Default) Models(ctx context.Context) (*agent.Models, error) {
	e := fmt.Sprintf("%s/%s", d.domain, _models_path)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e, nil)
	if err != nil {
		return nil, err
	}

	res, err := d.doRety(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var m agent.Models
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}
