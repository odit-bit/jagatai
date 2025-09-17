package cmd

import (
	"fmt"

	"github.com/odit-bit/jagatai/api"
	"github.com/spf13/cobra"
)

func init() {
	CliCompletionCMD.Flags().StringVar(&GlobEndpoint, "addr", "http://localhost:11823", "")
}

var (
	GlobEndpoint = ""
)

var CliCompletionCMD = cobra.Command{
	Use:  "chat args1",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		c := api.NewClient(GlobEndpoint, "")
		res, err := c.Chat(
			cmd.Context(),
			api.ChatRequest{
				Content: []*api.Message{
					api.NewTextMessage("user", args[0]),
				},
			})

		if err != nil {
			return err
		}

		fmt.Printf("> %s", res.Text)
		return nil
	},
}
