package main

import (
	"os"

	"github.com/kkrt-labs/kakarot-controller/cmd"
)

func main() {
	command := cmd.NewKKRTCtlCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
