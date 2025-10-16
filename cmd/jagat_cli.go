package cmd

import (
	"bufio"
	"bytes"
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
	Use:  "chat",
	Args: cobra.ExactArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {

		start(cmd.Context())

		return nil
	},
}

func start(ctx context.Context) {
	c := api.NewClient(GlobEndpoint, "")
	scanner := bufio.NewScanner(os.Stdin)
	session := session{}
	_ = session

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		input := scanner.Text()
		switch input {
		case "/exit":
			return
		}
		fmt.Printf("\n")
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

		fmt.Printf(">model: %s \n\n", res.Text)
		session.history = append(session.history, api.NewTextMessage("assistant", res.Text))
	}

}

type session struct {
	history []*api.Message
}

func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}
