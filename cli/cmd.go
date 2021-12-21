package cli

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	//chain "github.com/nickolation/fondness-chain/blockchain/chaincore"
	"github.com/nickolation/fondness-chain/fondcore/chain/assets"
	"github.com/nickolation/fondness-chain/fondcore/chain/chaincore"
	"github.com/nickolation/fondness-chain/fondcore/utils"
	"github.com/nickolation/fondness-chain/netcore/netfond"
	"github.com/spf13/cobra"
)

var (
	//errNilTx = errors.New("nil tx")
	errInvalidAddr = errors.New("invalid addr - cmd is locked")
	errNilNodeId   = errors.New("node_id isn't valid - cmd is locked")
)

var (
	idCursor = "NODE_ID"
)

//	Select NODE_ID from end
func FindId() (string, error) {
	id := os.Getenv(idCursor)

	if id == "" {
		return "", errNilNodeId
	}

	return id, nil
}

var minerAddr string

//	Start the node with node_id and mining flag setting.
//  If minerAddr isn't nil then node starts mining.
func InitNoder() {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "start the new node with NODE_ID .env value",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"hearter id",
				err,
			)

			if len(minerAddr) > 0 {
				if assets.ValidateAddr(minerAddr) {
					fmt.Printf("\nThe node is started in the miner mode. NODE_ID - [%s]\n", nodeId)
				}
				fmt.Printf("\nInvalid addr - mining is interrupt. NODE_ID - [%s]\n", nodeId)
			}
			netfond.StartNode(nodeId, minerAddr)
		},
	}

	cmd.Flags().StringVarP(
		&createAddr, "miner", "m", "",
		"The miner address is and string value",
	)

	CliRoot.AddCommand(cmd)
}

//	Print the chain cmd
func InitIndexer() {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "index utxo set",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {

			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"indexer id",
				err,
			)

			chain := chaincore.ExistChainStart(nodeId)
			defer chain.Db.Close()

			set := chaincore.UTXOSet{
				Chain: chain,
			}

			//	indexing the utxo set
			set.Index()

			ctr := set.CountUTX()
			fmt.Printf("UTXO Set is indexed! There are the [%v] utx\n", ctr)
		},
	}

	CliRoot.AddCommand(cmd)
}

//	Print the chain cmd
func InitPrinter() {
	cmd := &cobra.Command{
		Use:   "print",
		Short: "print chain blocks",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"printer id",
				err,
			)

			ch := chaincore.ExistChainStart(nodeId)
			iter := ch.Iterator()

			for iter.Step() {
				block := iter.Block()
				txn := iter.Txn()

				pow := chaincore.Pow(block)

				//	print info about blocks
				fmt.Println(block)
				fmt.Printf("  Pow valid - [%s]\n\n", strconv.FormatBool(pow.Validate()))

				//	print info about tx
				for _, tx := range txn {
					fmt.Println(tx)
				}
				//		Pow section
			}
		},
	}

	CliRoot.AddCommand(cmd)
}

//	Generate new heart and write it to the memory. Log the address
func InitHearter() {
	cmd := &cobra.Command{
		Use:   "heart",
		Short: "feel the heart - create and connect to",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"hearter id",
				err,
			)

			memory, err := assets.AccesMemory(nodeId)
			utils.Handle(
				"hearter err",
				err,
			)

			addr := memory.LinkHeart()
			memory.WriteMemory(nodeId)
			fmt.Printf("Address of heart is - [%s]\n", addr)
		},
	}

	CliRoot.AddCommand(cmd)
}

//	Print the list of the address nodes in the memory cmd
func InitLister() {
	cmd := &cobra.Command{
		Use:   "listaddr",
		Short: "list of all heart adresses in the memory",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"lister id",
				err,
			)

			memory, err := assets.AccesMemory(nodeId)
			utils.Handle(
				"hearter err",
				err,
			)

			list := memory.GetAddrs()
			for i, l := range list {
				fmt.Printf("Adress [%d] - [%s]\n", i, l)
			}
		},
	}

	CliRoot.AddCommand(cmd)
}

var createAddr string

//	Create new chain cmd
func InitCreator() error {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create a new fondness chain object at this addr",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"creator id",
				err,
			)

			chain := chaincore.AbsentChainStart(createAddr, nodeId)
			defer chain.Db.Close()

			set := chaincore.UTXOSet{
				Chain: chain,
			}
			set.Index()
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

//	Get balance by addr cmd
func InitBalancer() error {
	cmd := &cobra.Command{
		Use:   "fondness",
		Short: "print the level of fondness by this node address",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"balancer id",
				err,
			)

			if !assets.ValidateAddr(balanceAddr) {
				utils.FHandle(
					"CMD",
					errInvalidAddr,
				)
			}

			chain := chaincore.ExistChainStart(nodeId)
			defer chain.Db.Close()

			set := chaincore.UTXOSet{
				Chain: chain,
			}

			fmt.Printf(
				"\nFondness of loving by addr [%s] - [%d]\n",
				balanceAddr,
				set.GetFondness(balanceAddr),
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
	//	sender Node address
	fromAddr string
	//	getter Node address
	toAddr string

	//	amount of fondness
	force int

	//	node is miner status
	mineStatus string
)

//	Send fondness to the addr from loving cmd.
func InitLover() error {
	cmd := &cobra.Command{
		Use:   "love",
		Short: "send fondness from [Node] to [Node]",
		Long:  "...",
		Run: func(cmd *cobra.Command, args []string) {
			//	getting the node_id in this terminal
			nodeId, err := FindId()
			utils.FHandle(
				"lover id",
				err,
			)

			if !assets.ValidateAddr(toAddr) || !assets.ValidateAddr(fromAddr) {
				log.Fatal("addresses isn't valid")
			}

			chain := chaincore.ExistChainStart(nodeId)
			defer chain.Db.Close()

			set := chaincore.UTXOSet{
				Chain: chain,
			}

			mem, err := assets.AccesMemory(nodeId)
			utils.FHandle(
				"acces memory",
				err,
			)

			heart := mem.GetHeart(fromAddr)

			tx, err := set.ProduceTx(heart, toAddr, force)
			utils.Handle(
				"produce tx",
				err,
			)
			log.Printf("tx hash is - [%x]\n", tx.Hash)

			//	Generate the new tx with coinbase to the miner - sender
			if mineStatus == "on" {
				cbs := chaincore.CoinbaseTx(fromAddr, "")
				log.Printf("coinbase hash is - [%x]\n", cbs.Hash)
				if tx != nil {
					block := chain.Mine([]chaincore.Tx{*cbs, *tx})
					set.Refresh(block)
				}

				return
			}

			netfond.SendTX(netfond.ListNodes[0], tx)
			fmt.Printf(
				"\nTx is success! Fondness send\n[to] - [%s]\nin force - [%d]\n\n",
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
		&force, "force", "r", 0,
		"Force [Loving] - [Loving]",
	)

	cmd.Flags().StringVarP(
		&mineStatus, "mine", "m", "off",
		"Node [id] - is mine",
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

func init() {
	//	commands with the error returning
	utils.Handle(
		"creator",
		InitCreator(),
	)

	utils.Handle(
		"lover",
		InitLover(),
	)
	utils.Handle(
		"balancer",
		InitBalancer(),
	)

	//	unerrors commands
	InitPrinter()
	InitHearter()
	InitLister()
	InitIndexer()
	InitNoder()
}
