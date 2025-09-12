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
	ApiKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint"`
}

type AgentConfig struct {
	// Model string `yaml:"model"`
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
	if s, ok := os.LookupEnv("JAGATAI_PROVIDER_NAME"); ok {
		cfg.Provider.Name = s
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

	// llmAddr, err := flags.GetString(Flag_llm_address)
	// if err != nil {
	// 	mergedErr = errors.Join(mergedErr, err)
	// } else if llmAddr != "" {
	// 	cfg.Provider.Endpoint = llmAddr
	// }

	llmName, err := flags.GetString(Flag_llm_name)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if llmName != "" {
		cfg.Provider.Name = llmName
	}

	llmModel, err := flags.GetString(Flag_llm_model)
	if err != nil {
		mergedErr = errors.Join(mergedErr, err)
	} else if llmModel != "" {
		cfg.Provider.Model = llmModel
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

	return mergedErr
}
