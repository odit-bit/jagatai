package agent

import (
	"context"
	"fmt"
	"sync"
)

var driverMap = map[string]ProviderConstructFunc{}
var mutex = sync.Mutex{}

// Remote backend that serve model
type Provider interface {
	Chat(ctx context.Context, req CCReq) (*CCRes, error)
}

type ProviderConstructFunc func(key string) (Provider, error)

func RegisterDriver(name string, p ProviderConstructFunc) {
	mutex.Lock()
	defer mutex.Unlock()

	driverMap[name] = p
}

func NewWithProvider(model, driver, key string, opts ...OptionFunc) (*Agent, error) {
	p, ok := driverMap[driver]
	if !ok {
		return nil, fmt.Errorf("agent driver %s not found", driver)
	}

	adapter, err := p(key)
	if err != nil {
		return nil, err
	}
	return New(model, adapter, opts...), nil
}
