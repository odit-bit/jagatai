package tooldef

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/odit-bit/jagatai/agent"
)

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
	dmutex.Lock()
	defer dmutex.Unlock()

	t := []agent.Tool{}
	for _, cfg := range cfgs {
		fn, ok := providers[cfg.Name]
		if ok {
			p := fn(cfg)
			if !cfg.DisablePing {
				if ok, err := p.Ping(ctx); err != nil {
					return nil, fmt.Errorf("tools build: %s", err)
				} else if !ok {
					slog.Warn(fmt.Sprintf("tool build called but not available, maybe forget to register? Name: %s, Endpoint: %s", cfg.Name, cfg.Endpoint))
					continue //skip add the tool if ping fails
				}
			}

			//add tool
			t = append(t, p.Tooling())
			slog.Debug("tool initate", "name", cfg.Name, "address", cfg.Endpoint)
		}
	}
	if len(t) == 0 {
		slog.Warn("there is 0 tools provider registered, forget to blank import ? (toolprovider)")
	}
	return t, nil
}
