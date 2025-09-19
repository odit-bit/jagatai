package tooldef

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/odit-bit/jagatai/jagat/agent"
)

// managing tool life cycle

// configuration for 3rd party tool implementation
type Config struct {
	//name of tools that Register function use for discover
	Name        string
	//connection string for external call
	Endpoint    string
	//secret or api key for tool
	ApiKey      string
	//set true if tool need to make ping when it's build.
	//see agent.ToolProvider for interface.
	DisablePing bool
}

type ProviderConstructFunc func(cfg Config) agent.ToolProvider

var providers = make(map[string]ProviderConstructFunc)

var dmutex sync.RWMutex

func Register(name string, p ProviderConstructFunc) {
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

func Build(ctx context.Context, cfgs []Config) ([]agent.ToolProvider, error) {
	//temporary list provider
	type providerToBuild struct {
		provider agent.ToolProvider
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

	t := []agent.ToolProvider{}
	for _, item := range toBuild {
		if !item.config.DisablePing {
			if err := item.provider.Ping(ctx); err != nil {
				slog.Warn(
					fmt.Sprintf("skip build tool that not respond ping, Name: %s, Endpoint: %s",
						item.config.Name, item.config.Endpoint,
					))
				continue //skip add the tool
			}
		}

		//add tool
		t = append(t, item.provider)
		slog.Debug("tool initate", "name", item.config.Name, "address", item.config.Endpoint)
	}

	return t, nil
}

// RegisteredTools returns a list of all registered tool provider names.
func RegisteredTools() []string {
	dmutex.RLock()
	defer dmutex.RUnlock()
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	return names
}
