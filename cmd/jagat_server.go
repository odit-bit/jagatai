package cmd

import (
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
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		a, err := jagat.NewPflags(ctx, cmd.Flags())
		if err != nil {
			return err
		}

		return a.Run(ctx)
	},
}
