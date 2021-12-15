package chaincore

import (
	"bytes"

	"github.com/dgraph-io/badger"
)

//	Iterator struct for stepping the chain blocks
type ChainIterator struct {
	Db *badger.DB

	//	hash of current block
	Cursor []byte

	//	status of the end of chain
	//	is true if block == genesis
	Blocker bool

	//	block value for this cursor hash
	Value *FondBlock
}


//	Block-getter 
//	Return current block chain iteration epoch
func (iter *ChainIterator) Block () *FondBlock {
	return iter.Value
}


//	Txn-getter
func (iter *ChainIterator) Txn () []Tx {
	return iter.Value.Txn
}


//	Makes new iterator for stepping the chain
func (chain *FondChain) Iterator() *ChainIterator {
	return &ChainIterator{
		Db:     chain.Db,
		Cursor: chain.TailHash,
	}
}


//	Return true while the iterator is at the middle block
//	When iterator destinates the genesis it's false
func (iter *ChainIterator) Step() bool {
	var block *FondBlock

	if iter.Blocker {
		return false
	}
	
	err := iter.Db.View(func(txn *badger.Txn) error {
		tl, err := txn.Get(iter.Cursor)
		Handle(
			"getting current block",
			err,
		)

		err = tl.Value(func(val []byte) error {
			block = ToBlocker(val)
			return nil
		})

		return err
	})

	Handle(
		"getting block",
		err,
	)

	//	setting block value 
	iter.Value = block

	if block.IsGenesis() {
		iter.Blocker = true
		return true
	}

	if !bytes.Equal(block.PrevHash, []byte{}) {
		iter.Cursor = block.PrevHash
		return true
	}

	return false
}
