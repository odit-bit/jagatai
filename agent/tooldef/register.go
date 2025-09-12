package tooldef

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/odit-bit/jagatai/agent"
)

// managing tool life cycle

// configuration for 3rd party tool implementation
type Config struct {
	Name        string
	Endpoint    string
	ApiKey      string
	DisablePing bool
}

// represent tool provider
type Provider interface {
	Tooling() agent.Tool
	Ping(ctx context.Context) (bool, error)
}

type ProviderFunc func(cfg Config) Provider

var providers = make(map[string]ProviderFunc)

var dmutex sync.RWMutex

func Register(name string, p ProviderFunc) {
	dmutex.Lock()
	defer dmutex.Unlock()
	if p == nil {
		panic("tooldef: Register provider is nil")
	}
	if _, dup := providers[name]; dup {
		panic("tooldef: Register called twice for provider " + name)
	}
	providers[name] = p
}

func Count() int {
	dmutex.RLock()
	defer dmutex.RUnlock()
	return len(providers)
}

func Build(ctx context.Context, cfgs []Config) ([]agent.Tool, error) {
	//temporary list provider
	type providerToBuild struct {
		provider Provider
		config   Config
	}
	toBuild := []providerToBuild{}

	// --- Critical Section Start ---
	dmutex.RLock() // Use a Read Lock since we are only reading the map.
	for _, cfg := range cfgs {
		if fn, ok := providers[cfg.Name]; ok {
			p := fn(cfg)
			toBuild = append(toBuild, providerToBuild{provider: p, config: cfg})
		} else {
			slog.Warn("tool provider initated but not available, forget to register ?")
		}
	}
	dmutex.RUnlock()
	// --- Critical Section End --- Lock is now released.

	t := []agent.Tool{}
	for _, item := range toBuild {
		if !item.config.DisablePing {
			if ok, err := item.provider.Ping(ctx); err != nil {
				return nil, fmt.Errorf("tools build: %s", err)
			} else if !ok {
				slog.Warn(
					fmt.Sprintf("skip build tool that not respond ping, Name: %s, Endpoint: %s",
						item.config.Name, item.config.Endpoint,
					))
				continue //skip add the tool
			}
		}

		//add tool
		t = append(t, item.provider.Tooling())
		slog.Debug("tool initate", "name", item.config.Name, "address", item.config.Endpoint)
	}

	return t, nil
}
