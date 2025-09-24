package driver

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log/slog"
// 	"math"
// 	"math/rand"
// 	"net/http"
// 	"net/url"
// 	"strings"
// 	"time"

// 	"github.com/odit-bit/jagatai/jagat/agent"
// )

// const (
// 	_openai_domain           = "https://api.openai.com"
// 	_openai_completions_path = "v1/chat/completions"
// 	_models_path             = "v1/models"
// )

// const (
// 	_http_default_max_retry = 2
// )

// var _ agent.Provider = (*Default)(nil)

// // wrap simple OpenAI compatible api
// type Default struct {
// 	model  string
// 	hc     *http.Client
// 	apiKey string
// 	// domain    string
// 	baseUrl  string
// 	maxRetry int

// 	config Config
// }

// func NewOpenAIAdapter(model, key string, config *Config) (agent.Provider, error) {
// 	if model == "" {
// 		return nil, fmt.Errorf("openai_adapter model cannot be empty")
// 	}
// 	baseUrl, err := url.Parse(config.Endpoint)
// 	if err != nil {
// 		return nil, fmt.Errorf("openai_adapter failed parse base url: %w", err)
// 	}
// 	baseUrl.Path = ""

// 	return &Default{
// 		model: model,
// 		hc:    http.DefaultClient,
// 		// domain:    _openai_domain,
// 		apiKey:   key,
// 		maxRetry: _http_default_max_retry,
// 		baseUrl:  baseUrl.String(),
// 		config:   *config,
// 	}, nil
// }

// func (d *Default) doRety(req *http.Request) (*http.Response, error) {
// 	baseDelay := time.Duration(1 * time.Second)
// 	maxAttempt := d.maxRetry

// 	var errTry error
// 	for attempt := range maxAttempt {
// 		res, err := d.hc.Do(req)
// 		if err == nil && res.StatusCode < 400 {
// 			return res, nil
// 		}

// 		switch res.StatusCode {
// 		case 400, 404:
// 			return res, fmt.Errorf("driver failed send request status: %v", res.Status)
// 		}

// 		if res != nil {
// 			io.Copy(io.Discard, res.Body)
// 			res.Body.Close()
// 		}
// 		// exponential backoff with full jitter
// 		sleep := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt)))
// 		jitter := time.Duration(rand.Int63n(int64(sleep)))
// 		time.Sleep(jitter)
// 		slog.Debug("retry", "attempt", attempt, "max", maxAttempt, "status", res.StatusCode, "error", err)
// 		errTry = err
// 	}

// 	err1 := fmt.Errorf("api model request error: %v", errTry)
// 	err2 := fmt.Errorf("max attempt reach")
// 	errTry = errors.Join(err1, err2)

// 	return nil, errTry

// }

// // chat completions
// func (d *Default) Chat(ctx context.Context, req agent.CCReq) (*agent.CCRes, error) {

// 	input := Request{
// 		Model:       d.model,
// 		Messages:    toMessages(req.Messages),
// 		Temperature: *d.config.Temperature,
// 		TopP:        *d.config.TopP,
// 		Tools:       req.Tools,
// 		ToolChoice:  "auto",
// 	}

// 	b, err := json.Marshal(input)
// 	if err != nil {
// 		return nil, err
// 	}

// 	endpoint := fmt.Sprintf("%s/%s", d.baseUrl, _openai_completions_path)
// 	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewBuffer(b))
// 	if err != nil {
// 		return nil, err
// 	}

// 	slog.Debug("provider request", "endpoint", request.URL.String())
// 	request.Header.Set("Content-Type", "application/json")
// 	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.apiKey))

// 	resp, err := d.doRety(request)
// 	if err != nil {
// 		bb, _ := io.ReadAll(resp.Body)
// 		// jsonData, _ := json.MarshalIndent(req, "", "  ")
// 		// log.Println(string(jsonData))
// 		return nil, fmt.Errorf("openai adapter failed get response: %w  message: %s", err, string(bb))
// 	}
// 	defer resp.Body.Close()

// 	var out Response
// 	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
// 		return nil, fmt.Errorf("openai_adapter response error: %v", err)
// 	}

// 	tcalls := []*agent.ToolCall{}
// 	for _, tc := range out.Choices[0].Message.ToolCalls {
// 		tcalls = append(tcalls, &agent.ToolCall{
// 			ID:   tc.ID,
// 			Type: tc.Type,
// 			Function: agent.FunctionCall{
// 				Name:      tc.Function.Name,
// 				Arguments: tc.Function.Arguments,
// 			},
// 		})
// 	}

// 	// bb, _ := json.MarshalIndent(tcalls, "", " ")
// 	// log.Println("TOOLCAA:", string(bb))

// 	return &agent.CCRes{
// 		ID:     out.ID,
// 		Model:  out.Model,
// 		Object: out.Model,
// 		Choices: []agent.Choice{
// 			{
// 				Index:        out.Choices[0].Index,
// 				Text:         *out.Choices[0].Message.Content,
// 				ToolCalls:    tcalls,
// 				FinishReason: out.Choices[0].FinishReason,
// 			},
// 		},
// 	}, nil

// }

// type Request struct {
// 	Model       string       `json:"model"`
// 	Messages    []Message    `json:"messages"`
// 	Temperature float32      `json:"temperature,omitempty"`
// 	TopP        float32      `json:"top_p,omitempty"`
// 	Tools       []agent.Tool `json:"tools,omitempty"`
// 	ToolChoice  string       `json:"tool_choice,omitempty"`
// }

// func toMessages(src []*agent.Message) []Message {
// 	out := []Message{}
// 	for _, msg := range src {

// 		out = append(out, Message{
// 			Role:    string(msg.Role),
// 			Content: toParts(msg.Parts),
// 		})

// 	}
// 	return out
// }

// func toParts(src []*agent.Part) []Part {
// 	parts := []Part{}
// 	for _, in := range src {
// 		part := Part{}
// 		topart(in, &part)
// 		parts = append(parts, part)
// 	}
// 	return parts
// }

// func topart(in *agent.Part, out *Part) {
// 	// map text
// 	if in.Text != "" {
// 		out.Type = "text"
// 		out.Text = in.Text
// 		return
// 	}
// 	if in.ToolResponse != nil {
// 		out.Type = "text"
// 		out.Text = in.ToolResponse.String()
// 		return
// 	}

// 	// map blob
// 	if in.Blob != nil {
// 		if strings.Contains(in.Blob.Mime, "image") {
// 			buf := bytes.NewBufferString("data:")
// 			buf.WriteString(in.Blob.Mime)
// 			buf.WriteString(";base64,")
// 			buf.Write(in.Blob.Bytes)

// 			out.Type = "image_url"
// 			out.ImageURL = &ImageURL{
// 				URL: buf.String(),
// 			}
// 			return
// 		}
// 		return
// 	}
// }

// type Message struct {
// 	Role    string `json:"role"`
// 	Content []Part `json:"content"`
// }

// type Part struct {
// 	Type string `json:"type"` // "text" or "image_url"

// 	// Fields for a text block
// 	Text string `json:"text,omitempty"`

// 	// Fields for an image_url block
// 	ImageURL *ImageURL `json:"image_url,omitempty"`
// }

// type ImageURL struct {
// 	URL string `json:"url"`
// }

// //------------

// type Response struct {
// 	ID      string   `json:"id"`
// 	Object  string   `json:"object"`
// 	Created int64    `json:"created"`
// 	Model   string   `json:"model"`
// 	Choices []Choice `json:"choices"`
// }

// type Choice struct {
// 	Index        int          `json:"index"`
// 	Message      ResMessage   `json:"message"`
// 	Logprobs     *interface{} `json:"logprobs"` // null or object
// 	FinishReason string       `json:"finish_reason"`
// }

// type ResMessage struct {
// 	Role      string     `json:"role"`
// 	Content   *string    `json:"content"` // null or string
// 	ToolCalls []ToolCall `json:"tool_calls"`
// }

// type ToolCall struct {
// 	ID       string   `json:"id"`
// 	Type     string   `json:"type"`
// 	Function Function `json:"function"`
// }

// type Function struct {
// 	Name      string `json:"name"`
// 	Arguments string `json:"arguments"`
// }
