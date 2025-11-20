package jagat

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/driver"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
	_ "github.com/odit-bit/jagatai/jagat/agent/toolprovider"
)

type jagat struct {
	agent Agent
}

type Agent interface {
	Completion(ctx context.Context, msgs []*agent.Message) (*agent.Message, error)
}

func New(ctx context.Context, cfg *Config) (*jagat, error) {
	// Validate the final config
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	//logging
	if cfg.Server.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("configuration", "config", cfg)
	}

	// llm provider
	var provider agent.Provider
	var err error

	switch cfg.Provider.Name {

	case "ollama":
		provider, err = driver.NewOllamaAdapter(cfg.Provider.Model, cfg.Provider.ApiKey, &cfg.Provider.Options)

	case "genai":
		provider, err = driver.NewGeminiAdapter(cfg.Provider.Model, cfg.Provider.ApiKey, &cfg.Provider.Options)

	default:
		err = fmt.Errorf("unknown provider specified in config: %s", cfg.Provider.Name)

	}
	if err != nil {
		slog.Error("jagat init provider", "error", err)
		return nil, err
	}

	// Tools & MCP
	t, err := tooldef.Build(ctx, cfg.Tools)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}

	if cfg.MCP.Endpoint != "" {
		// add more tool via mcp
		tr, err := tooldef.NewMCPRegistry(ctx, cfg.MCP)
		if err != nil {
			return nil, err
		}

		t, err = tr.ListTools(ctx, t)
		if err != nil {
			return nil, err
		}
	}

	// Agent
	if cfg.Server.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("tools", "list", tooldef.RegisteredTools())
	}

	a := agent.New(
		provider,
		agent.WithTools(t...),
		agent.WithMaxToolCall(3),
	)

	return &jagat{
		agent: &a,
	}, nil
}

func (j *jagat) Completion(ctx context.Context, msgs []*agent.Message) (*agent.Message, error) {
	return j.agent.Completion(ctx, msgs)
}
