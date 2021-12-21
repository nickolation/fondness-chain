package chaincore

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/nickolation/fondness-chain/fondcore/utils"
)

var (
	Tail     = []byte("tl")
	stdForce = 1000
)

var (
	errUnverTx = errors.New("txn isn't verified")
)

//	main entitie of block is part of chain
type FondBlock struct {
	//	time adding
	Stamp int64

	//	idx of block in the chain - height
	Idx int

	//	block transactions
	Txn []Tx

	//	hash of this
	Hash []byte

	PrevHash []byte

	//	counter for pow functional
	Nonce int
}

//	produce new block before the linking with chain
func ProduceBlock(t []Tx, p []byte, idx int) FondBlock {
	block := FondBlock{
		Stamp:    time.Now().Unix(),
		Idx:      idx,
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

//	Mine the block and set it to the db
func (chain *FondChain) Mine(txn []Tx) *FondBlock {
	var (
		tailHash []byte
		tailIdx  int
	)

	for _, tx := range txn {
		if !chain.VefifyTX(&tx) {
			log.Printf("hash is - [%x]\n", tx.Hash)
			utils.FHandle(
				"unverified tx - mine is locked",
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

		it, err := txn.Get(tailHash)
		utils.FHandle(
			"getting tail block item",
			err,
		)

		var block []byte
		err = it.Value(func(val []byte) error {
			block = val
			return nil
		})
		utils.FHandle(
			"getting tail block value",
			err,
		)

		b := ToBlocker(block)
		tailIdx = b.Idx

		return err
	})

	Handle(
		"view last hash",
		err,
	)

	block := ProduceBlock(txn, tailHash, tailIdx+1)

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

//	Check the block is in the base.
//	Set block in the base by hash.
//	If block.Idx > tail.Idx -> update the tail: tailHash = block.hash.
//	It provides the making of the best height chain in blockhain.
func (chain *FondChain) LinkBlock(block *FondBlock) {
	err := chain.Db.Update(func(txn *badger.Txn) error {

		//	Block already is in the base
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockBuff := block.ToByter()
		err := txn.Set(block.Hash, blockBuff)
		utils.FHandle(
			"set the block by hash",
			err,
		)

		it, err := txn.Get([]byte(Tail))
		utils.FHandle(
			"get the tail block",
			err,
		)

		var tailHash []byte
		err = it.Value(func(val []byte) error {
			tailHash = val
			return nil
		})
		utils.FHandle(
			"get the tail data-hash",
			err,
		)

		b, err := txn.Get(tailHash)
		utils.FHandle(
			"get the block data by last hash",
			err,
		)

		var blockBytes []byte
		err = b.Value(func(val []byte) error {
			blockBytes = val
			return nil
		})
		utils.FHandle(
			"get the block data - serialization",
			err,
		)

		tail := ToBlocker(blockBytes)

		if block.Idx > tail.Idx {
			err = txn.Set([]byte(Tail), block.Hash)
			utils.FHandle(
				"set the new tail",
				err,
			)
			chain.TailHash = block.Hash
		}

		return nil
	})
	utils.Handle(
		"update the tail with new block",
		err,
	)
}

//	Returns the max size of blockchain in the base.
func (chain *FondChain) MaxSize() int {
	var tail FondBlock

	err := chain.Db.View(func(txn *badger.Txn) error {
		it, err := txn.Get([]byte(Tail))
		utils.FHandle(
			"get the tail block",
			err,
		)

		var tailHash []byte
		err = it.Value(func(val []byte) error {
			tailHash = val
			return nil
		})
		utils.FHandle(
			"get the tail data-hash",
			err,
		)

		b, err := txn.Get(tailHash)
		utils.FHandle(
			"get the block data by last hash",
			err,
		)

		var blockBytes []byte
		err = b.Value(func(val []byte) error {
			blockBytes = val
			return nil
		})
		utils.FHandle(
			"get the block data - serialization",
			err,
		)

		tail = *ToBlocker(blockBytes)
		return nil
	})
	utils.Handle(
		"get tail block for the idx value",
		err,
	)

	return tail.Idx
}

//	Get the matrix of all blocks hashes in the database.
func (chain *FondChain) BlocksHashes() [][]byte {
	var matrix [][]byte

	iter := chain.Iterator()

	for iter.Step() {
		matrix = append(matrix, iter.Block().Hash)
	}

	return matrix
}

//	Search the block by hash in the badgere database.
func (chain *FondChain) BlockByHash(hash []byte) (FondBlock, error) {
	var block FondBlock

	err := chain.Db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(hash)
		utils.FHandle(
			"get the block-data by hash",
			err,
		)

		var buff []byte
		err = it.Value(func(val []byte) error {
			buff = val
			return nil
		})
		utils.FHandle(
			"get the data of item - block",
			err,
		)

		block = *ToBlocker(buff)
		return nil
	})
	utils.FHandle(
		"get the block by hash",
		err,
	)

	return block, err
}

//	Generate the genesis block with coinbase tx
func FondGenesis(coinbase Tx) *FondBlock {
	block := ProduceBlock([]Tx{coinbase}, []byte{}, 0)
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
func ExistChainStart(id string) *FondChain {
	path := fmt.Sprintf(dbPath, id)
	if !DbExist(path) {
		log.Printf("Chain [db] isn't exist")
		runtime.Goexit()
	}

	var tailHash []byte

	opts := badger.DefaultOptions(path)

	db, err := connDb(path, opts)
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

//	Init chain without db. Start the chain and write in in the badger.
func AbsentChainStart(addr string, id string) *FondChain {
	path := fmt.Sprintf(dbPath, id)
	if DbExist(path) {
		log.Printf("Chain [db] is already exist")
		runtime.Goexit()
	}

	var tailHash []byte

	opts := badger.DefaultOptions(path)

	db, err := connDb(path, opts)
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
