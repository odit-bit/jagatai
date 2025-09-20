package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/labstack/echo/v4"
	"github.com/odit-bit/jagatai/jagat"
	"github.com/odit-bit/jagatai/jagat/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

func init() {
	ServerCMD.Flags().AddFlagSet(config.FlagSet)
}

var ServerCMD = cobra.Command{
	Use:  "run",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		// config file
		cfg, err := config.LoadAndValidate(cmd.Flags())
		if err != nil {
			slog.Error("invalid configuration", "error", err)
			return fmt.Errorf("invalid configuration: %w", err)
		}

		a, err := jagat.New(ctx, cfg)
		if err != nil {
			return err
		}

		// http server
		e := echo.New()
		e.Debug = cfg.Server.Debug
		e.GET("/metric", echo.WrapHandler(promhttp.Handler()))
		jagat.RestHandler(ctx, a, e)

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
