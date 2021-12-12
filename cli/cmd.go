package cli

import (
	"fmt"
	"log"
	"strconv"

	//chain "github.com/nickolation/fondness-chain/blockchain/chaincore"
	"github.com/nickolation/fondness-chain/blockchain/chaincore"
	"github.com/nickolation/fondness-chain/core/utils"
	"github.com/spf13/cobra"
)

var (
//errNilTx = errors.New("nil tx")
)

func (cli *CliChain) InitPrinter() {
	ch := chaincore.ExistChainStart("")

	iter := ch.Iterator()

	cmd := &cobra.Command{
		Use:   "print",
		Short: "print chain blocks",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			for iter.Step() {
				val := iter.Val()

				pow := chaincore.Pow(val)
				fmt.Printf("Prev hash: [%x] \n", val.PrevHash)
				fmt.Printf("Hash: [%x] \n", val.Hash)
				fmt.Printf("Pow valid - [%s]\n\n", strconv.FormatBool(pow.Validate()))
				//		Pow section
			}
		},
	}

	CliRoot.AddCommand(cmd)
}

var createAddr string

func (cli *CliChain) InitCreator() error {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a new fondness chain object at this addr",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			chaincore.AbsentChainStart(createAddr)

			fmt.Println("Fondness chain is created!")
		},
	}

	cmd.Flags().StringVarP(
		&createAddr, "addr", "a", "",
		"The data of the address",
	)

	err := cmd.MarkFlagRequired("addr")
	CliRoot.AddCommand(cmd)

	return err
}

var balanceAddr string

func (cli *CliChain) InitBalancer() error {
	cmd := &cobra.Command{
		Use:   "fondness",
		Short: "print the level of fondness by this node address",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			ch := chaincore.ExistChainStart(balanceAddr)

			fmt.Printf(
				"Fondness of loving by addr [%s] - [%d]\n",
				balanceAddr,
				ch.GetFondness(balanceAddr),
			)
		},
	}

	cmd.Flags().StringVarP(
		&balanceAddr, "addr", "a", "",
		"The data of the address",
	)

	err := cmd.MarkFlagRequired("addr")
	CliRoot.AddCommand(cmd)

	return err
}

var (
	fromAddr string
	toAddr   string

	force int
)

func (cli *CliChain) InitLover() error {
	cmd := &cobra.Command{
		Use:   "love",
		Short: "send fondness from [Node] to [Node]",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			ch := chaincore.ExistChainStart(fromAddr)

			tx, err := ch.ProduceTx(fromAddr, toAddr, force)
			utils.Handle(
				"produce tx",
				err,
			)
			log.Printf("Tx is - [%v]", tx)

			if tx != nil {
				ch.LinkBlock([]chaincore.Tx{*tx})
				fmt.Printf("CMD: tx isn't nil - [%v]\n", tx)
			}

			fmt.Printf(
				"Tx is success! Fondness send from [%s]\n to - [%s]\n - in force [%d]\n",
				fromAddr,
				toAddr,
				force,
			)
		},
	}

	cmd.Flags().StringVarP(
		&fromAddr, "from", "f", "",
		"Sender [Loving] - [Loving]",
	)

	cmd.Flags().StringVarP(
		&toAddr, "to", "t", "",
		"Get [Loving] - [Loving]",
	)

	cmd.Flags().IntVarP(
		&force, "force", "frc", 0,
		"Force [Loving] - [Loving]",
	)

	err := cmd.MarkFlagRequired("from")
	utils.Handle(
		"data flag",
		err,
	)

	err = cmd.MarkFlagRequired("to")
	utils.Handle(
		"to flag",
		err,
	)

	err = cmd.MarkFlagRequired("force")
	utils.Handle(
		"force flag",
		err,
	)

	CliRoot.AddCommand(cmd)

	return err
}
