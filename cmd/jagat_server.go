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
	FLAG_PROVIDER_KEY      = "provider.apikey"
	FLAG_PROVIDER_ENDPOINT = "provider.endpoint"
	FLAG_PROVIDER_NAME     = "provider.name"
	FLAG_PROVIDER_MODEL    = "provider.model"

	FLAG_SERVER_ADDRESS     = "server.address"
	FLAG_SERVER_DEBUG       = "server.debug"
	FLAG_SERVER_CONFIG_FILE = "config"

	FLAG_METRIC_PROMETHEUS = "metric.prometheus"
	FLAG_TRACE_JAEGER      = "trace.jaeger"
)

// Defined set of flags for jagat configuration use.
var FlagSet = pflag.NewFlagSet("Jagat_Flags", pflag.PanicOnError)

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
	FlagSet.Bool(FLAG_METRIC_PROMETHEUS, false, "enable default prometheus")
	FlagSet.Bool(FLAG_TRACE_JAEGER, false, "enable default jaeger exporter")
}

// load configuration from default embedded config.yaml, provided config.yaml, env and flags before validation.
func LoadAndValidate(flags *pflag.FlagSet) (*jagat.Config, error) {
	v := viper.New()
	v.SetConfigType("yaml")

	// Bind env variable
	v.SetEnvPrefix("JAGATAI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind Pflags flags
	v.BindPFlags(flags)

	// Set default value by reading from the embedded config.yaml
	defaultBytes, _ := defaultConfig.ReadFile("jagat_server_config.yaml")
	if err := v.ReadConfig(bytes.NewReader(defaultBytes)); err != nil {
		return nil, fmt.Errorf("failed to read default config: %w", err)
	}

	// Set from external config file if provided
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
