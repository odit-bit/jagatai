package tooldef

import (
	"context"

	"github.com/odit-bit/jagatai/agent"
)

type Config struct {
	Name        string
	Endpoint    string
	ApiKey      string
	DisablePing bool
}

type Provider interface {
	Tooling() agent.Tool
	Ping(ctx context.Context) (bool, error)
}

type ProviderFunc func(cfg Config) Provider
