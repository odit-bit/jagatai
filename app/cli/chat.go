package main

import (
	"fmt"

	"github.com/odit-bit/jagatai/client"
	"github.com/spf13/cobra"
)

func init() {
	CompletionsCMD.Flags().StringVar(&GlobEndpoint, "addr", "http://localhost:11823", "")
}

var (
	GlobEndpoint = ""
)

var CompletionsCMD = cobra.Command{
	Use:  "chat args1",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		c := client.NewClient(GlobEndpoint, "")
		res, err := c.Chat(client.ChatRequest{
			Messages: []client.Message{
				{
					Role: "user",
					Text: args[0],
				},
			},
		})

		if err != nil {
			return err
		}

		fmt.Printf("> %s", res.Message.Text)
		return nil
	},
}
