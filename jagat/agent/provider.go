package agent

import (
	"context"
)

// var driverMap = map[string]ProviderConstructFunc{}
// var mutex = sync.Mutex{}

// Remote backend that serve model
type Provider interface {
	Chat(ctx context.Context, req CCReq) (*CCRes, error)
}

// type ProviderConstructFunc func(key string) (Provider, error)

// func RegisterDriver(name string, p ProviderConstructFunc) {
// 	mutex.Lock()
// 	defer mutex.Unlock()

// 	driverMap[name] = p
// }

// // DriverRegistry holds all available provider constructors.
// type DriverRegistry map[string]ProviderConstructFunc

// func NewWithProvider(model, driver, key string, registry DriverRegistry, opts ...OptionFunc) (*Agent, error) {

// 	p, ok := registry[driver]
// 	if !ok {
// 		return nil, fmt.Errorf("agent driver %s not found", driver)
// 	}

// 	adapter, err := p(key)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return New(model, adapter, opts...), nil
// }
