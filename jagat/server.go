package jagat

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/odit-bit/jagatai/jagat/agent"
	"github.com/odit-bit/jagatai/jagat/agent/driver"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
	_ "github.com/odit-bit/jagatai/jagat/agent/toolprovider"
	"github.com/spf13/pflag"
)

type jagat struct {
	address string
	Agent
}

type Agent interface {
	Completion(ctx context.Context, msgs []*agent.Message) (*agent.Message, error)
}

func new(ctx context.Context, cfg *Config) (*jagat, error) {
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
		Agent:   a,
		address: cfg.Server.Address,
	}, nil
}

func (j *jagat) Address() string {
	return j.address
}

// create jagat instance from flags
func NewPflags(ctx context.Context, flag *pflag.FlagSet) (*jagat, error) {
	// config file
	cfg, err := LoadAndValidate(flag)
	if err != nil {
		slog.Error("invalid configuration", "error", err)
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// otel
	// Initialize observability
	shutdown := InitObservability(ctx, "jagat-server", cfg.Observe)
	defer shutdown(ctx) // Ensure shutdown is called on exit
	a, err := new(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (j *jagat) Run(ctx context.Context) error {
	// http server
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	e := echo.New()
	RestHandler(ctx, j.Agent, e)

	var err error
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- e.Start(j.address)
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		stop()
	}

	return e.Shutdown(ctx)
}
