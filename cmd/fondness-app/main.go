package main

import (
	"os"
	"github.com/nickolation/fondness-chain/cli"
	"github.com/nickolation/fondness-chain/core/utils"
)

func main() {
	defer os.Exit(0)

	utils.Handle(
		"MAIN: cli err",
		cli.InitCli(),
	)
}
