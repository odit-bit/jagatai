package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/odit-bit/jagatai/jagat"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	defineFlags()
	ServerCMD.Flags().AddFlagSet(FlagSet)
}

//go:embed jagat_server_config.yaml
var defaultConfig embed.FS

const (
	FLAG_PROVIDER_KEY      = "p_key"
	FLAG_PROVIDER_ENDPOINT = "p_addr"
	FLAG_PROVIDER_NAME     = "p_name"
	FLAG_PROVIDER_MODEL    = "p_model"

	FLAG_SERVER_ADDRESS     = "addr"
	FLAG_SERVER_DEBUG       = "debug"
	FLAG_SERVER_CONFIG_FILE = "config"

	FLAG_OBSERVE_ENABLE         = "observe"
	FLAG_OBSERVE_TRACEENDPOINT  = "traceendpoint"
	FLAG_OBSERVE_METER_ENDPOINT = "metricendpoint"
)

// Defined set of flags for jagat configuration use.
var FlagSet = pflag.NewFlagSet("Jagat_Flags", pflag.PanicOnError)

var flagToConfigKeyMap = map[string]string{
	FLAG_PROVIDER_KEY:      "provider.apikey",
	FLAG_PROVIDER_ENDPOINT: "provider.endpoint",
	FLAG_PROVIDER_NAME:     "provider.name",
	FLAG_PROVIDER_MODEL:    "provider.model",

	FLAG_SERVER_ADDRESS: "server.address",
	FLAG_SERVER_DEBUG:   "server.debug",
	// FLAG_SERVER_CONFIG_FILE: "config",

	FLAG_OBSERVE_ENABLE:        "observe.enable",
	FLAG_OBSERVE_TRACEENDPOINT: "",
}

func defineFlags() {
	// server
	FlagSet.String(FLAG_SERVER_ADDRESS, "", "server address")
	FlagSet.Bool(FLAG_SERVER_DEBUG, false, "debug log")
	FlagSet.String(FLAG_SERVER_CONFIG_FILE, "", "path to config file")

	// provider
	FlagSet.String(FLAG_PROVIDER_KEY, "", "provider's api key")
	FlagSet.String(FLAG_PROVIDER_NAME, "", "provider's name")
	FlagSet.String(FLAG_PROVIDER_MODEL, "", "provider's model name")

	//observe
	FlagSet.Bool(FLAG_OBSERVE_ENABLE, false, "enable observability default false")
}

// load configuration from default embedded config.yaml, provided config.yaml, env and flags before validation.
func LoadAndValidate(flags *pflag.FlagSet) (*jagat.Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// 3. Bind env variable
	v.SetEnvPrefix("JAGATAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 4. Bind Pflags flags
	for flagName, configKey := range flagToConfigKeyMap {
		v.BindPFlag(configKey, flags.Lookup(flagName))
	}

	// 1. Set default value by reading from the embedded config.yaml
	defaultBytes, _ := defaultConfig.ReadFile("jagat_server_config.yaml")
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
			return nil, fmt.Errorf("failed to read provided config: %w", err)
		}
	}

	// 5. UNmarshal
	var cfg jagat.Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

var ServerCMD = cobra.Command{
	Use:  "server",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		// jagat config
		cfg, err := LoadAndValidate(cmd.Flags())
		if err != nil {
			slog.Error(err.Error())
			// return err
		}

		j, err := jagat.NewHttp(ctx, *cfg)
		if err != nil {
			slog.Error(err.Error())
			// return err
		}

		go func() {
			<-ctx.Done()
			slog.Debug("received shutdown signal")
		}()

		if err := j.Start(); err != nil {
			slog.Error(err.Error())
		}

	},
}
