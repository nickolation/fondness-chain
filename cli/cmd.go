package cli

import (
	"fmt"

	//chain "github.com/nickolation/fondness-chain/blockchain/chaincore"
	"github.com/spf13/cobra"
)

func (cli *CliChain) InitPrinter() {
	iter := cli.Chain.Iterator()

	cmd := &cobra.Command{
		Use:   "print",
		Short: "print chain blocks",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			for iter.Step() {
				val := iter.Val()

				fmt.Printf("Prev hash: [%x] \n", val.PrevHash)
				fmt.Printf("Hash: [%x] \n", val.Hash)
				fmt.Printf("Data: [%x] \n", val.Data)

				//		Pow section
			}
		},
	}

	CliRoot.AddCommand(cmd)
}

var blockData string

func (cli *CliChain) InitLinker() error {
	cmd := &cobra.Command{
		Use:   "link",
		Short: "link the block to fondness chain",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Chain.LinkBlock([]byte(blockData))
			fmt.Println("Link the new block to the chain")
		},
	}

	cmd.Flags().StringVarP(
		&blockData, "data", "d", "",
		"The data of the chain block",
	)

	err := cmd.MarkFlagRequired("data")
	CliRoot.AddCommand(cmd)

	return err
}
