package jagat

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/driver"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
	_ "github.com/odit-bit/jagatai/jagat/agent/toolprovider"
	"github.com/odit-bit/jagatai/jagat/config"
)

type jagat struct {
	Agent
}

type Agent interface {
	Completion(ctx context.Context, msgs []*agent.Message) (*agent.Message, error)
}

func New(ctx context.Context, cfg *config.Config) (*jagat, error) {
	//logging
	if cfg.Server.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("configuration", "config", cfg)
	}
	// provider
	var provider agent.Provider
	var err error

	switch cfg.Provider.Name {
	case "ollama":
		provider, err = driver.NewOllamaAdapter(cfg.Provider.Model, cfg.Provider.ApiKey, &cfg.Provider.Extra)
	// case "openai":
	// 	provider, err = driver.NewOpenAIAdapter(cfg.Provider.Model, cfg.Provider.ApiKey, &cfg.Provider.Extra)
	case "genai":
		provider, err = driver.NewGeminiAdapter(cfg.Provider.Model, cfg.Provider.ApiKey, &cfg.Provider.Extra)

	default:
		err = fmt.Errorf("unknown provider specified in config: %s", cfg.Provider.Name)

	}
	if err != nil {
		slog.Error("jagat init provider", "error", err)
		return nil, err
	}

	// tools
	t, err := tooldef.Build(ctx, cfg.Tools)
	if err != nil {
		slog.Error(err.Error())
		return nil, err
	}
	toolOpt := agent.WithTool(t...)
	if cfg.Server.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("tools", "list", tooldef.RegisteredTools())
	}

	// agent
	a := agent.New(provider, toolOpt)

	return &jagat{
		Agent: a,
	}, nil
}
