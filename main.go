package main

import (
	"github.com/odit-bit/jagatai/cmd"
	"github.com/spf13/cobra"
)

func main() {
	rootCMD := cobra.Command{}
	rootCMD.AddCommand(
		&cmd.ServerCMD, 
		&cmd.CliCompletionCMD, 
		&cmd.TeleCMD,
	)
	rootCMD.Execute()
}
