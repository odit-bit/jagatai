package main

import (
	"log"

	"github.com/odit-bit/jagatai/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCMD := cobra.Command{}
	rootCMD.AddCommand(
		&cmd.ServerCMD,
		&cmd.TeleCMD,
		&cmd.CliCompletionCMD,
	)
	if err := rootCMD.Execute(); err != nil {
		log.Println(err)
	}
}
