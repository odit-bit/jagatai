package client

import (
	"bytes"
	"encoding/json"
	"errors"
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
}

func NewClient(endpoint, key string) *Client {
	return &Client{
		client:   http.DefaultClient,
		Endpoint: endpoint,
	}
}

func (c *Client) Chat(in ChatRequest) (*ChatResponse, error) {
	const path = "v1/chat/completions"
	urlString := fmt.Sprintf("%s/%s", c.Endpoint, path)

	b, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Post(urlString, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out ChatResponse
	if resp.StatusCode > 299 {
		b, _ := io.ReadAll(resp.Body)
		// fmt.Printf("> error read body: %s, status: %d", string(b), resp.StatusCode)
		return nil, errors.New(string(b))
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, err
		}
	}

	return &out, nil
}
