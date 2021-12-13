package main

import (
	"os"
	"github.com/nickolation/fondness-chain/cli"
)

func main() {
	defer os.Exit(0)
	cli.Execute()
}
