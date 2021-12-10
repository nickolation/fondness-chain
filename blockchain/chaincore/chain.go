package chaincore

import (
	"log"

	"github.com/dgraph-io/badger"
)

//		[]byte --> ?
var (
	genesis = []byte("genesis")
	Tail = []byte("tl")

	sourcePath = "./source/chain"
)

//	main entitie of block is part of chain 
type FondBlock struct {
	//		data --> tx
	Data []byte 
	
	//	hash of this 
	Hash []byte 

	PrevHash []byte 

	//	counter for pow functional
	Nonce int
}

//	produce new block before the linking with chain
func ProduceBlock(d, p []byte) FondBlock {
	block := FondBlock{
		Data: d,
		PrevHash: p,
	}

	//	init temp 
	pow := NewPow(block)
	h, n := pow.Feel()

	block.Hash = h 
	block.Nonce = n

	return block
}


//	main etitie of chain 
type FondChain struct {
	Db *badger.DB
	TailHash []byte
} 


//	init the block with data
//	link new block with fondChain 
func (chain *FondChain) LinkBlock(d []byte) {
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

	block := ProduceBlock(d, tailHash)

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


//	start fondchain with genesis block
func StartChain() *FondChain {
	var tailHash []byte 

	opt := badger.DefaultOptions(sourcePath)

	db, err := badger.Open(opt)
	Handle(
		"error with open the chain base",
		err,
	)

	//	try to update db
	err = db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(Tail); err == badger.ErrKeyNotFound {
			log.Println("Fondness chain isn't exist")

			gen := ProduceBlock(genesis, nil)

			log.Printf("genesis is - [%v]", gen)
			//	set genesis block
			Handle(
				"error with setting genesis block to base",
				txn.Set(gen.Hash, gen.ToByter()),
			)

			//set tail
			err = txn.Set(Tail, gen.Hash)
			tailHash = gen.Hash

			Handle(
				"error with setting tail",
				err,
			)
			return err

		//	something is here
		} else {
			tl, err := txn.Get(Tail)
			Handle(
				"getting tail",
				err,
			)

			Handle(
				"getting tail value",
				tl.Value(func(val []byte) error {
					tailHash = val
					return nil
				}),
			)
			
			return err
		}
	})

	Handle(
		"updating base",
		err,
	)

	return &FondChain{
		TailHash: tailHash,
		Db: db,
	}
}



