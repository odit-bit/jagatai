package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Message struct {
	Role    string
	Content string
}

type ChatRequest struct {
	Messages []Message
}

type ChatResponse struct {
	Message    Message
	DoneReason string
	CreatedAt  time.Time
}

type Client struct {
	client   *http.Client
	Endpoint string
	key      string
}

func NewClient(endpoint, key string) *Client {
	return &Client{
		client:   http.DefaultClient,
		Endpoint: endpoint,
		key:      key,
	}
}

func (c *Client) Chat(in ChatRequest) (*ChatResponse, error) {
	const path = "v1/chat/completions"
	urlString := fmt.Sprintf("%s/%s", c.Endpoint, path)

	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, urlString, bytes.NewReader(b))
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
