package cmd

import (
	"log/slog"
	"os"
	"os/signal"

	"github.com/odit-bit/jagatai/jagat"
	"github.com/spf13/cobra"
)

func init() {
	ServerCMD.Flags().AddFlagSet(jagat.FlagSet)
}

var ServerCMD = cobra.Command{
	Use:  "server",
	Args: cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		// jagat config
		cfg, err := jagat.LoadAndValidate(cmd.Flags())
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
