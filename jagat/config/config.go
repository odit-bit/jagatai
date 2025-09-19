package config

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/odit-bit/jagatai/jagat/agent/driver"
	"github.com/odit-bit/jagatai/jagat/agent/tooldef"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//go:embed config.yaml
var defaultConfig embed.FS

// holds aggregats configuration across jagat environment.
type Config struct {
	Server   ServerConfig     `yaml:"server"`
	Provider Provider         `yaml:"provider"`
	Tools    []tooldef.Config `yaml:"tools"`
}

// jagat server config
type ServerConfig struct {
	Address string `yaml:"address"`
	Debug   bool   `yaml:"debug"`
}

// external llm provider
type Provider struct {
	Name     string        `yaml:"name"`
	Model    string        `yaml:"model"`
	ApiKey   string        `yaml:"apikey"`
	Endpoint string        `yaml:"endpoint"`
	Extra    driver.Config `yaml:"extra"`
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

	if c.Provider.Model == "" {
		return errors.New("provider model is required")
	}

	return nil
}

// load configuration from default embedded config.yaml, provided config.yaml, env and flags before validation.
func LoadAndValidate(flags *pflag.FlagSet) (*Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// 3. Bind env variable
	v.SetEnvPrefix("JAGATAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. Bind Pflags flags
	for flagName, configKey := range flagToConfigKeyMap {
		// if err := v.BindPFlag(configKey, flags.Lookup(flagName)); err != nil {
		// 	return nil, fmt.Errorf("failed to bind flags %s:%w", flagName, err)
		// }
		v.BindPFlag(configKey, flags.Lookup(flagName))
	}

	// 1. Set default value by reading from the embedded config.yaml
	defaultBytes, _ := defaultConfig.ReadFile("config.yaml")
	if err := v.ReadConfig(bytes.NewReader(defaultBytes)); err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	// 2.Set from external config file if provided
	configFile, _ := flags.GetString(FLAG_SERVER_CONFIG_FILE)
	if configFile != "" {
		f, err := os.Open(configFile)
		if err != nil {
			return nil, fmt.Errorf("config : %w", err)
		}
		defer f.Close()
		providedBytes, _ := io.ReadAll(f)
		if err := v.ReadConfig(bytes.NewReader(providedBytes)); err != nil {
			return nil, fmt.Errorf("failed to read default config: %w", err)
		}
	}

	// 5. UNmarshal
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. Validate the final config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}
