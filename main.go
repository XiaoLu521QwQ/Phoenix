package main

import (
	"github.com/pterm/pterm"
	"phoenix/minecraft"
)

// The following program implements a proxy that forwards players from one local address to a remote address.
func main() {
	pterm.Info.Println("Fast Builder Lambda: Starting ... ")
	pterm.Info.Println("Author: CAIMEO")
	minecraft.Run("config.toml")
}

