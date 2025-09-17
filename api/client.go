package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	default_address = "http://127.0.0.1:11823"
)

type Client struct {
	client   *http.Client
	Endpoint string
	key      string
}

func NewClient(endpoint, key string) *Client {
	if endpoint == "" {
		endpoint = default_address
	}
	return &Client{
		client:   http.DefaultClient,
		Endpoint: endpoint,
		key:      key,
	}
}

func (c *Client) Chat(ctx context.Context, in ChatRequest) (*ChatResponse, error) {
	const path = "v1/chat/completions"
	urlString := fmt.Sprintf("%s/%s", c.Endpoint, path)

	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlString, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("client failed create request: %v", err)
	}

	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("Authorization", fmt.Sprintf("Bearer %s", c.key))

	req.Header = header

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out ChatResponse
	if resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		// fmt.Printf("> error read body: %s, status: %d", string(b), resp.StatusCode)
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(b))
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
	}

	return &out, nil
}
