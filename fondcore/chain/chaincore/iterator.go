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
func (iter *ChainIterator) Block() *FondBlock {
	return iter.Value
}

//	Txn-getter
func (iter *ChainIterator) Txn() []Tx {
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

//	Iterator object for the iterating keys in badger db.
//	Ctr is auto incrementing in the time pushing to the keys
type KeysIterator struct {
	//	Prefixed hash-tx key matrix
	Keys [][]byte

	//	Iterator's switcher
	Ctr int
}

//	Keys host constructor
func KeysIter() *KeysIterator {
	return &KeysIterator{
		Keys: make([][]byte, 0, utxosetSize),
	}
}

//	Push the key in the keys storage.
func (host *KeysIterator) Push(key []byte) {
	host.Keys = append(host.Keys, key)
	host.Ctr++
}

//	Validation on the finish point destination
func (host *KeysIterator) IsFinished() bool {
	return host.Ctr == utxosetSize
}

//	Validation on the permission to delete the prefixed keys
func (host *KeysIterator) IsFree() bool {
	return host.Ctr > 0
}

//	Reconstuct of the iterator with zero-value fields
func (host *KeysIterator) Vanish() {
	host.Keys = make([][]byte, 0, utxosetSize)
	host.Ctr = 0
}
