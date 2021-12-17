package netfond

import (
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
	deathChan := make(chan os.Signal)

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

	cmd := DecodeCmd(req)
	log.Printf("\ncmd is - [%x]\n", cmd)

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
