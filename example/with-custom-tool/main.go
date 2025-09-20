package main

import (
	"log"

	"github.com/odit-bit/jagatai/cmd"

	// Import your custom tool package with a blank import.
	// This is the crucial step to ensure your tool's init() function is called.
	_ "github.com/odit-bit/jagatai/example/with-custom-tool/osm"
)

func main() {

	// run jagat as rest server using predefined flags and command.
	if err := cmd.ServerCMD.Execute(); err != nil {
		log.Println(err)
	}

}
