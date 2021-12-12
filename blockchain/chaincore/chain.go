package chaincore

import (
	"log"
	"runtime"

	"github.com/dgraph-io/badger"
)

//		[]byte --> ?
var (
	genesis = []byte("genesis")
	Tail    = []byte("tl")

	sourcePath = "./source/chain"

	stdForce = 1000
)

//	main entitie of block is part of chain
type FondBlock struct {
	//	block transactions
	Txn []Tx

	//	hash of this
	Hash []byte

	PrevHash []byte

	//	counter for pow functional
	Nonce int
}

//	produce new block before the linking with chain
func ProduceBlock(t []Tx, p []byte) FondBlock {
	block := FondBlock{
		Txn:      t,
		PrevHash: p,
	}

	//	init temp
	pow := Pow(&block)
	h, n := pow.Feel()

	block.Hash = h
	block.Nonce = n

	return block
}

//	main etitie of chain
type FondChain struct {
	Db       *badger.DB
	TailHash []byte
}

//	init the block with data
//	link new block with fondChain
func (chain *FondChain) LinkBlock(t []Tx) {
	var tailHash []byte

	//	getting tail hash
	err := chain.Db.View(func(txn *badger.Txn) error {
		tl, err := txn.Get(Tail)
		Handle(
			"getting tail",
			err,
		)

		err = tl.Value(func(val []byte) error {
			tailHash = val
			return nil
		})

		Handle(
			"getting tail value",
			err,
		)

		return err
	})

	Handle(
		"view last hash",
		err,
	)

	block := ProduceBlock(t, tailHash)

	//	update chain - add new block and write tail to base
	err = chain.Db.Update(func(txn *badger.Txn) error {
		Handle(
			"setting new block to base",
			txn.Set(block.Hash, block.ToByter()),
		)

		err = txn.Set(Tail, block.Hash)
		Handle(
			"setting new tail",
			err,
		)
		return err
	})

	Handle(
		"updating chain - add new block",
		err,
	)
}

//	Generate the genesis block with coinbase tx
func FondGenesis(coinbase Tx) *FondBlock {
	block := ProduceBlock([]Tx{coinbase}, nil)
	return &block
}

//	Validate block in the genesis parametrs
func (block *FondBlock) IsGenesis() bool {
	if block.PrevHash == nil && block.Txn[0].IsCoinbase() {
		return true
	}

	return false
}

//		init chain with db
func ExistChainStart(addr string) *FondChain {
	if !DbExist(dbPath) {
		log.Printf("Chain [db] isn't exist")
		runtime.Goexit()
	}

	var tailHash []byte

	opt := badger.DefaultOptions(sourcePath)
	db, err := badger.Open(opt)
	Handle(
		"error with open the chain base",
		err,
	)

	err = db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(Tail)
		Handle(
			"getting the tail",
			err,
		)

		err = it.Value(func(val []byte) error {
			tailHash = val
			return nil
		})

		return err
	})

	Handle(
		"getting the tail value",
		err,
	)

	return &FondChain{
		Db:       db,
		TailHash: tailHash,
	}
}

//		init chain without db
func AbsentChainStart(addr string) *FondChain {
	if DbExist(dbPath) {
		log.Printf("Chain [db] is already exist")
		runtime.Goexit()
	}

	var tailHash []byte

	opt := badger.DefaultOptions(sourcePath)
	db, err := badger.Open(opt)
	Handle(
		"error with open the chain base",
		err,
	)

	err = db.Update(func(txn *badger.Txn) error {

		cbsTx := CoinbaseTx(addr, string(genesis))
		gen := FondGenesis(*cbsTx)

		//	uncamp log
		log.Printf("genesis is - [%v]", gen)

		Handle(
			"setting the tail",
			txn.Set(Tail, gen.Hash),
		)

		err = txn.Set(gen.Hash, gen.ToByter())
		Handle(
			"setting the new block",
			err,
		)

		tailHash = gen.Hash

		return err
	})

	Handle(
		"updating the chain",
		err,
	)

	return &FondChain{
		Db:       db,
		TailHash: tailHash,
	}
}
