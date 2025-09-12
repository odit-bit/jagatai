package middleware

import (
	"fmt"
	"sync"

	"github.com/odit-bit/jagatai/agent"
)

var (
	registry = map[string]MiddlewareConstructFunc{}
	mutex    = sync.Mutex{}
)

type Config map[string]any

type MiddlewareConstructFunc func(name string, config Config) (agent.MiddlewareFunc, error)

func Build(name string, config Config) (agent.MiddlewareFunc, error) {
	m, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("middleware %v is not found, forget to register ?", name)
	}
	return m(name, config)
}

func Register(name string, fn MiddlewareConstructFunc) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := registry[name]; ok {
		panic(fmt.Sprintf("register %v twice", name))
	}
	registry[name] = fn
}
