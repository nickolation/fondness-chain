package chaincore

import (
	"encoding/hex"
	"errors"
	"log"
	"runtime"

	"github.com/dgraph-io/badger"
	"github.com/nickolation/fondness-chain/core/utils"
)


var (
	Tail    = []byte("tl")

	sourcePath = "./source/chain"

	stdForce = 1000
)

var (
	errUnverTx = errors.New("txn isn't verified")
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
		Hash:     []byte{},
		Txn:      t,
		PrevHash: p,
		Nonce:    0,
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
func (chain *FondChain) LinkBlock(txn []Tx) *FondBlock {
	var tailHash []byte

	for _, tx := range txn {
		if !chain.VefifyTX(&tx) {
			utils.FHandle(
				"unverified tx - link is locked",
				errUnverTx,
			)
		}
	}

	//	getting tail hash
	err := chain.Db.View(func(txn *badger.Txn) error {
		tl, err := txn.Get(Tail)
		utils.FHandle(
			"getting tail",
			err,
		)

		err = tl.Value(func(val []byte) error {
			tailHash = val
			return nil
		})

		utils.FHandle(
			"getting tail value",
			err,
		)

		return err
	})

	Handle(
		"view last hash",
		err,
	)

	block := ProduceBlock(txn, tailHash)

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

	return &block
}

//	Generate the genesis block with coinbase tx
func FondGenesis(coinbase Tx) *FondBlock {
	block := ProduceBlock([]Tx{coinbase}, []byte{})
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

		cbsTx := CoinbaseTx(addr, "")
		gen := FondGenesis(*cbsTx)

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


//	Iterates the chain and founds all utx - utxo map elements.
//	Veiws it in the map [hash] - []OutTx sort	 
func (chain *FondChain) ViewUTXO() map[string]TXOSet {
	//	mapping the parent tx hash and txoSet - slie of outs
	utxo := make(map[string]TXOSet)
	//	spent tx
	stx := make(map[string][]int) 

	iter := chain.Iterator() 
	for iter.Step() {
		txn := iter.Txn() 

		for _, tx := range txn {
			txHash := hex.EncodeToString(tx.Hash)

		Out:
			//	check outId of current tx to spent by map of spentable idx
			for outIdx, out := range tx.Out {
				if stx[txHash] != nil {
					for _, sId := range stx[txHash] {
						if sId == outIdx {
							continue Out
						}
					}
				}
				
				//	appending the out slice to map by hash of parent tx
				outs := utxo[txHash]
				outs.Outs = append(outs.Outs, out)
				utxo[txHash] = outs
			}

			//	search spentable tx
			//	coinbase isn't valid for stx map appending
			if !tx.IsCoinbase() {
				for _, in := range tx.In {
					rTxHash := hex.EncodeToString(in.Ref)
					stx[rTxHash] = append(stx[rTxHash], in.RefIdx)
				}
			}
		}
	}

	return utxo
}