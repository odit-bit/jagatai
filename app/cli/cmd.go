package main

import (
	"github.com/spf13/cobra"
)

func main() {
	rootCMD := cobra.Command{}
	rootCMD.AddCommand(&CompletionsCMD)
	rootCMD.Execute()
}
