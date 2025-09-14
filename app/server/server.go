package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/odit-bit/jagatai/agent"
	"github.com/odit-bit/jagatai/agent/driver"
	_ "github.com/odit-bit/jagatai/agent/driver"
	"github.com/odit-bit/jagatai/agent/middleware"
	"github.com/odit-bit/jagatai/agent/tooldef"
	_ "github.com/odit-bit/jagatai/agent/toolprovider"
	"github.com/odit-bit/jagatai/app/server/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func init() {
	ServerCMD.Flags().String(config.Flag_srv_addr, "", "server address")
	ServerCMD.Flags().Bool(config.Flag_srv_debug, false, "debug log")
	ServerCMD.Flags().String(config.Flag_srv_configfile, "", "path to config file")

	ServerCMD.Flags().String(config.Flag_llm_key, "", "provider's api key")
	ServerCMD.Flags().String(config.Flag_llm_name, "", "provider's name")
	ServerCMD.Flags().String(config.Flag_llm_model, "", "base model agent use")

}

var ServerCMD = cobra.Command{
	Use:  "run model",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		// config file
		cfg := config.Config{}
		configFilepath, _ := cmd.Flags().GetString(config.Flag_srv_configfile)
		if err := config.UnmarshalConfigFile(configFilepath, &cfg); err != nil {
			slog.Error(err.Error())
			return err
		}
		if err := config.BindConfigEnv(&cfg); err != nil {
			slog.Error(err.Error())
			return err
		}

		if err := config.BindCobraFlags(&cfg, cmd.Flags()); err != nil {
			slog.Error(err.Error())
			return err
		}

		// // Handle shutdown properly so nothing leaks.
		// otelShutdown, err := setupOtel(ctx)
		// if err != nil {
		// 	logger.Error(err.Error())
		// 	return err
		// }
		// defer func() {
		// 	err = errors.Join(err, otelShutdown(context.Background()))
		// }()
		// reg, err := updateRAMUsage()
		// if err != nil {
		// 	return err
		// }
		// defer reg.Unregister()

		//logging
		if cfg.Server.Debug {
			slog.SetLogLoggerLevel(slog.LevelDebug)
			slog.Debug("configuration", "config", cfg)
		}

		// provider
		var provider agent.Provider
		var err error

		switch cfg.Provider.Name {
		case "openai":
			provider, err = driver.NewOpenAIAdapter(cfg.Provider.ApiKey)
		case "genai":
			provider, err = driver.NewGenaiAdapter(cfg.Provider.ApiKey)
		default:
			err = fmt.Errorf("unknown provider specified in config: %s", cfg.Provider.Name)
		}
		if err != nil {
			slog.Error("failed to create agent provider", "error", err)
			return err
		}

		// tools
		t, err := tooldef.Build(ctx, cfg.Tools)
		if err != nil {
			slog.Error(err.Error())
			return err
		}
		toolOpt := agent.WithTool(t...)

		if err != nil {
			slog.Error("jagatAI init provider", "error", err.Error())
			return err
		}
		// agent
		a, err := agent.NewPipe(
			cfg.Provider.Model,
			provider,
			toolOpt,
		)
		if err != nil {
			slog.Error("jagatAI init Agent", "error", err.Error())
			return err
		}
		// middleware
		m, _ := middleware.Build("mock", middleware.Config{})
		a.AddMiddleware(m)

		// http server
		e := echo.New()
		e.Debug = cfg.Server.Debug
		e.GET("/metric", echo.WrapHandler(promhttp.Handler()))
		HandleAgent(ctx, a, e)

		srvErr := make(chan error, 1)
		go func() {
			srvErr <- e.Start(cfg.Server.Address)
		}()

		select {
		case err = <-srvErr:
			return err
		case <-ctx.Done():
			stop()
		}

		return e.Shutdown(ctx)
	},
}
