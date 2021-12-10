package cli

import (
	"log"

	chain "github.com/nickolation/fondness-chain/blockchain/chaincore"
)

type CliChain struct {
	Chain *chain.FondChain
}

func InitCli(c *chain.FondChain) error {
	cch := CliChain{
		Chain: c,
	}

	cch.InitPrinter()
	
	err := cch.InitLinker()
	if err != nil {
		log.Printf(
			"link add command - [%v]",
			err,
		)
	}

	Execute()
	return err
}