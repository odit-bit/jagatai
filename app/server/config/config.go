package config

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"strings"

	"github.com/odit-bit/jagatai/agent/tooldef"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	Flag_llm_key = "p_key"
	// Flag_llm_address = "p_addr"
	Flag_llm_name  = "p_name"
	Flag_llm_model = "p_model"

	Flag_srv_addr       = "addr"
	Flag_srv_debug      = "debug"
	Flag_srv_configfile = "config"
)

//go:embed config.yaml
var defaultConfig embed.FS
var (
	_default_config_file = "config.yaml"
)

type Config struct {
	Server   ServerConfig     `yaml:"server"`
	Agent    AgentConfig      `yaml:"agent"`
	Provider Provider         `yaml:"provider"`
	Tools    []tooldef.Config `yaml:"tools"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
	Debug   bool   `yaml:"debug"`
}

type Provider struct {
	Name     string `yaml:"name"`
	Model    string `yaml:"model"`
	ApiKey   string `yaml:"key"`
	Endpoint string `yaml:"endpoint"`
}

type AgentConfig struct {
	// Model string `yaml:"model"`
}

// Validate checks the configuration for correctness.
func (c *Config) Validate() error {
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

	// add a check for supported providers
	supportedProviders := map[string]bool{"ollama": true, "openai": true, "genai": true, "genai_rest": true}
	if !supportedProviders[c.Provider.Name] {
		return fmt.Errorf("unsupported provider: %s", c.Provider.Name)
	}
	switch c.Provider.Name {
	case "genai", "openai":
		if c.Provider.ApiKey == "" {
			return fmt.Errorf("provider chosen need api key: %v", c.Provider.Name)
		}
	}

	if c.Provider.Model == "" {
		return errors.New("provider model is required")
	}

	return nil
}

func LoadAndValidate(flags *pflag.FlagSet) (*Config, error) {
	v := viper.New()

	// 1. Set default value by reading from the embedded config.yaml
	defaultBytes, _ := defaultConfig.ReadFile("config.yaml")
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(defaultBytes)); err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	// 2.Set from external config file if provided
	configFile, _ := flags.GetString(Flag_srv_configfile)
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.MergeInConfig(); err != nil {
			if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
				return nil, fmt.Errorf("failed to merge config file: %w", err)
			}
		}
	}

	// 3. Bind env variable
	v.SetEnvPrefix("JAGATAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. Bind Pflags flags
	v.BindPFlag("server.address", flags.Lookup(Flag_srv_addr))
	v.BindPFlag("server.debug", flags.Lookup(Flag_srv_debug))
	v.BindPFlag("provider.apikey", flags.Lookup(Flag_llm_key))
	v.BindPFlag("provider.name", flags.Lookup(Flag_llm_name))
	v.BindPFlag("provider.model", flags.Lookup(Flag_llm_model))

	// 5. UNmarshal
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. Validate the final config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	slog.Debug("Config")
	return &cfg, nil
}
