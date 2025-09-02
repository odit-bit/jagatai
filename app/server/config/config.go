package config

import (
	"embed"
	"errors"
	"fmt"
	"os"

	"github.com/odit-bit/jagatai/agent/tooldef"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

const (
	Flag_llm_key     = "llm_key"
	Flag_llm_address = "llm_addr"

	Flag_srv_addr       = "addr"
	Flag_srv_debug      = "debug"
	Flag_srv_configfile = "config"

	Flag_agent_model = "model"
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
	ApiKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint"`
}

type AgentConfig struct {
	Model string `yaml:"model"`
}

func UnmarshalConfigFile(filename string, cfg *Config) error {
	var b []byte
	var err error
	if filename != "" {
		b, err = os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("config: %s", err)
		}
	} else {
		b, err = defaultConfig.ReadFile(_default_config_file)
		if err != nil {
			return fmt.Errorf("config: %s", err)
		}
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return fmt.Errorf("config unmarshal: %s", err)
	}

	return nil
}

func BindConfigEnv(cfg *Config) error {
	// get from environment
	if s, ok := os.LookupEnv("JAGATAI_PROVIDER_ENDPOINT"); ok {
		cfg.Provider.Endpoint = s
	}
	if s, ok := os.LookupEnv("JAGATAI_PROVIDER_APIKEY"); ok {
		cfg.Provider.ApiKey = s
	}

	return nil
}

func BindCobraFlags(cfg *Config, flags *pflag.FlagSet) error {

	var mergedErr error

	//llm
	llmKey, err := flags.GetString(Flag_llm_key)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if llmKey != "" {
		cfg.Provider.ApiKey = llmKey
	}

	llmAddr, err := flags.GetString(Flag_llm_address)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if llmAddr != "" {
		cfg.Provider.Endpoint = llmAddr
	}

	//server
	srvAddr, err := flags.GetString(Flag_srv_addr)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if srvAddr != "" {
		cfg.Server.Address = srvAddr
	}

	srvDebug, err := flags.GetBool(Flag_srv_debug)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if srvDebug {
		cfg.Server.Debug = srvDebug
	}

	// agent
	model, err := flags.GetString(Flag_agent_model)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if model != "" {
		cfg.Agent.Model = model
	}

	return mergedErr
}
