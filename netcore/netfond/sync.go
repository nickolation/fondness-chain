package netfond

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/nickolation/fondness-chain/fondcore/chain/chaincore"
	"github.com/nickolation/fondness-chain/netcore/utils"
)

func DeadNode(chain *chaincore.FondChain) {
	deathChan := make(chan os.Signal, 2)

	signal.Notify(deathChan, syscall.SIGTERM, syscall.SIGINT)
	<-deathChan

	err := chain.Db.Close()
	utils.FHandle(
		"db close",
		err,
	)

	defer os.Exit(1)
	defer runtime.Goexit()
}


//	--> chain method
func HandleConn(conn net.Conn, chain *chaincore.FondChain) {
	req, err := ioutil.ReadAll(conn)
	defer conn.Close()

	utils.FHandle(
		"request getting from conn",
		err,
	)

	cmd := DecodeCmd(req[:cmdClaim])
	log.Printf("\ncmd is - [%s]\n", cmd)

	//	handle the all commands
	switch cmd {
	case "addr":
		HandleAddr(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		log.Println("Unknown command")
	}

}


// Start the Node server with listening the localhost:id addr and sending the miner info
func StartNode(id string, minerAddr string) {
	syncAddr = fmt.Sprintf("localhost:%s", id)

	log.Printf("node addr is - [%s]", syncAddr)
	mineAddr = minerAddr 

	ln, err := net.Listen(protocol, syncAddr)
	utils.FHandle(
		"start the node server",
		err,
	)

	defer ln.Close() 

	chain := chaincore.ExistChainStart(id)
	go DeadNode(chain)

	sync := ListNodes[0]
	if syncAddr != sync {
		SendVersion(sync, chain)
	}

	for {
		conn, err := ln.Accept()
		utils.FHandle(
			"connection to the node",
			err,
		)

		go HandleConn(conn, chain)
	}
}