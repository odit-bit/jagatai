package jagat

import (
	"errors"
	"fmt"
	"net"

	"github.com/odit-bit/jagatai/jagat/agent/driver"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
)

// holds aggregats configuration across jagat environment.
type Config struct {
	Server   ServerConfig      `yaml:"server"`
	Provider Provider          `yaml:"provider"`
	Tools    []tooldef.Config  `yaml:"tools"`
	MCP      tooldef.MCPConfig `yaml:"mcp"`
	// Observe  ObsConfig        `yaml:"observability"`
	Metric Metric
	Trace  Trace
}

// jagat server config
type ServerConfig struct {
	Address string `yaml:"address"`
	Debug   bool   `yaml:"debug"`
}

// external llm provider
type Provider struct {
	Name     string
	Model    string
	ApiKey   string
	Endpoint string
	Options  driver.Config
}

type ObsConfig struct {
	Enable bool
	// if not set but enable will use stdout
	Exporter string
	// http endpoint exporter
	TraceEndpoint string
	//
	MetricsEndpoint string
	// secure endpoint (https)
	Secure bool
}

// Validate checks the configuration for correctness.
func (c *Config) validate() error {
	if c.Server.Address == "" {
		return errors.New("server address is required")
	}
	// Check if the address is a valid host:port
	if _, _, err := net.SplitHostPort(c.Server.Address); err != nil {
		return fmt.Errorf("invalid server address format: %w", err)
	}

	if c.Provider.Name == "" {
		return errors.New("provider name is required")
	}

	if c.Provider.Model == "" {
		return errors.New("provider model is required")
	}

	return nil
}
