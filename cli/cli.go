package cli

import "github.com/nickolation/fondness-chain/core/utils"

type CliChain struct {
}

func InitCli() error {
	cch := CliChain{}

	
	cch.InitPrinter()
	utils.Handle(
		"balancer",
		cch.InitBalancer(),
	)

	utils.Handle(
		"lover",
		cch.InitLover(),
	)
	
	utils.Handle(
		"creator",
		cch.InitCreator(),
	)

	Execute()

	//	nil --> err
	return nil
}
