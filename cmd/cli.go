package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"

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
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {

		// c := api.NewClient(GlobEndpoint, "")
		// res, err := c.Chat(
		// 	cmd.Context(),
		// 	api.ChatRequest{
		// 		Content: []*api.Message{
		// 			api.NewTextMessage("user", args[0]),
		// 		},
		// 	})

		// if err != nil {
		// 	fmt.Printf("> %s", err)
		// 	return nil
		// }

		// fmt.Printf("> %s", res.Text)
		// return nil

		start(cmd.Context())

		return nil
	},
}

func start(ctx context.Context) {
	c := api.NewClient(GlobEndpoint, "")
	scanner := bufio.NewScanner(os.Stdin)
	session := session{}

	for scanner.Scan() {
		input := scanner.Text()
		switch input {
		case "/exit":
			return
		}
		session.history = append(session.history, api.NewTextMessage("user", input))

		res, err := c.Chat(
			ctx,
			api.ChatRequest{
				Content: session.history,
			})

		if err != nil {
			fmt.Printf(">error: %s \n", err)
			return
		}

		fmt.Printf(">model: %s \n", res.Text)
		session.history = append(session.history, api.NewTextMessage("assistant", res.Text))
	}

}

type session struct {
	history []*api.Message
}
