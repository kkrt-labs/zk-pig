package main

import (
	"os"

	"github.com/kkrt-labs/zk-pig/cmd"
)

func main() {
	command := cmd.NewZkPigCommand()
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
