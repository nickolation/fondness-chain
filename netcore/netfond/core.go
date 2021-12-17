package netfond

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"io"
	"log"
	"net"

	"github.com/nickolation/fondness-chain/fondcore/chain/chaincore"
	"github.com/nickolation/fondness-chain/netcore/utils"
)

const (
	version  = 1
	cmdClaim = 12

	syncInit = "localhost:3000"
)

var (
	syncAddr string
	mineAddr string

	listNodes = []string{syncInit}

	blockTransfer = [][]byte{}

	txTransfer = make(map[string]chaincore.Tx)
)

type Addr struct {
	List []string
}

type Block struct {
	From string

	Block []byte
}

type GetBlocks struct {
	From string
}

type GetData struct {
	From string

	Kind string

	Hash []byte
}

type Inv struct {
	From string

	Kind string

	Pool [][]byte
}

type Tx struct {
	From string

	Tx []byte
}

type Version struct {
	From string

	Version int
	MaxSize int
}

var (
	protocol = "tcp"
)

func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)
	defer conn.Close()

	if err != nil {
		utils.Log(
			"addr isn't available",
			err,
		)

		var clearNodes []string
		for _, a := range listNodes {
			if a != addr {
				clearNodes = append(clearNodes, a)
			}
		}

		listNodes = clearNodes
	}

	_, err = io.Copy(conn, bytes.NewReader(data))
	utils.FHandle(
		"copy data to conn is interrupt",
		err,
	)
}

//	Send Addrs data to the address
func SendAddr(addr string) {
	a := Addr{
		List: listNodes,
	}

	a.List = append(a.List, syncAddr)
	tail := Encode(a)

	req := append(EncodeCmd("addr"), tail...)
	SendData(addr, req)
}

//	Send Block data to the address
func SendBlock(addr string, block *chaincore.FondBlock) {
	b := Block{
		From:  syncAddr,
		Block: block.ToByter(),
	}

	tail := Encode(b)

	req := append(EncodeCmd("block"), tail...)
	SendData(addr, req)
}

//	Send Inv data to the address
func SendInv(addr string, kind string, value [][]byte) {
	i := Inv{
		From: syncAddr,
		Kind: kind,
		Pool: value,
	}

	tail := Encode(i)

	req := append(EncodeCmd("inv"), tail...)
	SendData(addr, req)
}

//	Send GetBlocks data to the address
func SendGetBlocks(addr string) {
	gb := GetBlocks{
		From: syncAddr,
	}

	tail := Encode(gb)

	req := append(EncodeCmd("getblocks"), tail...)
	SendData(addr, req)
}

//	Send GetData data to the address
func SendGetData(addr string, kind string, hash []byte) {
	gd := GetData{
		From: syncAddr,
		Kind: kind,
		Hash: hash,
	}

	tail := Encode(gd)

	req := append(EncodeCmd("getdata"), tail...)
	SendData(addr, req)
}

//	Send Tx data to the address
func SendTX(addr string, tx *chaincore.Tx) {
	t := Tx{
		From: syncAddr,
		Tx:   tx.ToByte(),
	}

	tail := Encode(t)

	req := append(EncodeCmd("tx"), tail...)
	SendData(addr, req)
}

//	Send Version data to the address
func SendVersion(addr string, chain *chaincore.FondChain) {
	max := chain.MaxSize()
	v := Version{
		From:    syncAddr,
		Version: version,
		MaxSize: max,
	}

	tail := Encode(v)

	req := append(EncodeCmd("version"), tail...)
	SendData(addr, req)
}

func HandleAddr(request []byte) {
	var buff bytes.Buffer
	var addr Addr

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write addr data to bytes",
		err,
	)

	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to addr",
		dec.Decode(&addr),
	)

	listNodes = append(listNodes, addr.List...)
	log.Printf("THE LIST COUNT - [%d]", len(listNodes))
	RequestBlocks()
}

//	Send addr data to the blocks by GetBlocks cmd
func RequestBlocks() {
	for _, addr := range listNodes {
		SendGetBlocks(addr)
	}
}

func HandleBlock(request []byte, chain *chaincore.FondChain) {
	var buff bytes.Buffer
	var block Block

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write block data to bytes",
		err,
	)

	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to block",
		dec.Decode(&block),
	)

	b := chaincore.ToBlocker(block.Block)
	log.Println("\nGet the new block!\n")

	//	next logic
	chain.AddBlock(b)
	log.Printf("\nAdded the new block with hash - [%x]!\n", b.Hash)

	if len(blockTransfer) > 0 {
		bHash := blockTransfer[0]
		SendGetData(block.From, "block", bHash)

		blockTransfer = blockTransfer[1:]
	} else {
		UTXOSet := chaincore.UTXOSet{chain}
		UTXOSet.Index()
	}
}


func HandleGetBlocks(request []byte, chain *chaincore.FondChain) {
	var buff bytes.Buffer
	var blocks GetBlocks

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write GetBlocks data to bytes",
		err,
	)

	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to blocks",
		dec.Decode(&blocks),
	)


	//	next logic 
	hashes := chain.GetBlockHashes()
	SendInv(blocks.From, "block", hashes)
}


func HandleGetData(request []byte, chain *chaincore.FondChain) {
	var buff bytes.Buffer
	var data GetData

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write GetData data to bytes",
		err,
	)

	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to blocks",
		dec.Decode(&data),
	)


	kind := data.Kind
	hash := data.Hash
	addr := data.From

	switch kind {
	case "block":
		//	next logic 
		b, err := chain.GetBlock([]byte(hash))	
		utils.FHandle(
			"getting the block by hash",
			err,
		)

		SendBlock(addr, b)
	case "tx":
		txHash := hex.EncodeToString(hash)
		tx := txTransfer[txHash]

		SendTX(addr, &tx)
	}
}

func HandleVersion(request []byte, chain *chaincore.FondChain) {
	var buff bytes.Buffer
	var v Version

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write version data to bytes",
		err,
	)
	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to version",
		dec.Decode(&v),
	)


	//	next logic 
	maxSize := chain.GetBestHeight()
	otherSize := v.MaxSize

	if maxSize < otherSize {
		SendGetBlocks(v.From)
	}

	//	alternative suit 
	SendVersion(v.From, chain)


	if !NodeExistance(v.From) {
		listNodes = append(listNodes, v.From)
	}
}


func HandleTx(request []byte, chain *chaincore.FondChain) {
	var buff bytes.Buffer
	var T Tx

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write tx data to bytes",
		err,
	)
	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to tx",
		dec.Decode(&T),
	)
	
	tx := chaincore.ToTX(T.Tx)
	txTransfer[hex.EncodeToString(tx.Hash)] = tx

	log.Printf("%s, %d", syncAddr, len(txTransfer))

	if syncAddr == listNodes[0] {
		for _, addr := range listNodes {
			if addr != syncAddr && addr != T.From {
				SendInv(addr, "tx", [][]byte{tx.Hash})
			}
		}
	}

	if len(txTransfer) >= 2 && len(mineAddr) > 0 {
		MineTx(chain)
	}
}


func MineTx(chain *chaincore.FondChain) {
	var txs []*chaincore.Tx

	for hash := range txTransfer {
		log.Printf("tx: %s\n", txTransfer[hash].Hash)

		tx := txTransfer[hash]
		if chain.VefifyTX(&tx) {
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		log.Println("All Transactions are invalid")
		return
	}

	cbTx := chaincore.CoinbaseTx(mineAddr, "")
	txs = append(txs, cbTx)

	//	next logic 
	newBlock := chain.MineBlock(txs)

	UTXOSet  := chaincore.UTXOSet{chain}
	UTXOSet.Index()

	log.Println("New Block mined")

	for _, tx := range txs {
		txHash := hex.EncodeToString(tx.Hash)
		delete(txTransfer, txHash)
	}

	for _, addr := range listNodes {
		if addr != syncAddr {
			SendInv(addr, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(txTransfer) > 0 {
		MineTx(chain)
	}
}


func HandleInv(request []byte, chain *chaincore.FondChain) {
	var buff bytes.Buffer
	var inv Inv

	_, err := buff.Write(request[cmdClaim:])
	utils.FHandle(
		"write inv data to bytes",
		err,
	)
	dec := gob.NewDecoder(&buff)
	utils.FHandle(
		"decode bytes to inv",
		dec.Decode(&inv),
	)

	log.Printf("Recevied inventory with %d %s\n", len(inv.Pool), inv.Kind)


	kind := inv.Kind 
	addr := inv.From

	switch kind {
	case "block":
		blockHash := inv.Pool[0]

		SendGetData(addr, "block", blockHash)

		newTransfer := [][]byte{}
		transfer := inv.Pool
		for _, b := range transfer {
			if bytes.Compare(b, blockHash) != 0 {
				newTransfer = append(newTransfer, b)
			}
		}

		transfer = newTransfer

	case "tx":	
		txHash := inv.Pool[0]

		if txTransfer[hex.EncodeToString(txHash)].Hash == nil {
			SendGetData(addr, "tx", txHash)
		}
	}
}