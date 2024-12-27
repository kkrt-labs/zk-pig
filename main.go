package main

import (
	"github.com/kkrt-labs/kakarot-controller/cmd"
)

func main() {
	// Runs the root command which (has version + prover-inputs subcommands)
	cmd.Execute()
}
