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
	// address string
	Agent
}

type Agent interface {
	Completion(ctx context.Context, msgs []*agent.Message) (*agent.Message, error)
}

func New(ctx context.Context, cfg *Config) (*jagat, error) {
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
		// address: cfg.Server.Address,
	}, nil
}

// // create jagat instance from flags
// func NewPflags(ctx context.Context, flag *pflag.FlagSet) (*jagat, error) {
// 	// config file
// 	cfg, err := LoadAndValidate(flag)
// 	if err != nil {
// 		slog.Error("invalid configuration", "error", err)
// 		return nil, fmt.Errorf("invalid configuration: %w", err)
// 	}

// 	// otel
// 	// Initialize observability
// 	shutdown := InitObservability(ctx, "jagat-server", cfg.Observe)
// 	defer shutdown(ctx) // Ensure shutdown is called on exit
// 	a, err := New(ctx, cfg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return a, nil
// }
