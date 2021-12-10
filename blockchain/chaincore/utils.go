package chaincore

import (
	"bytes"
	"encoding/gob"
	"log"

	"github.com/dgraph-io/badger"
)

//	Hanlder with logger based on getting description
func Handle(des string, err error) {
	if err != nil {
		log.Printf("%s - [%v]", des, err)
	}
}

//	FondBlock to []byte
func (b *FondBlock) ToByter() []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)

	Handle(
		"error with encode block to byte",
		enc.Encode(b),
	)

	return buff.Bytes()
}

//	[]byte to FondBlock
func ToBlocker(s []byte) *FondBlock {
	var block FondBlock
	d := gob.NewDecoder(bytes.NewReader(s))

	Handle(
		"error with decoding byte to block",
		d.Decode(&block),
	)

	return &block
}

//	iterator struct for stepping the chain blocks
type ChainIterator struct {
	Db *badger.DB

	//	hash of current block
	Cursor []byte

	//	pointer of the end of chain
	//	is false if block == genesis
	Blocker bool

	//	block value for this cursor hash
	Value *FondBlock
}

//	val getter 
//	return current block chain iteration epoch
func (iter *ChainIterator) Val () *FondBlock {
	return iter.Value
}

//	makes new iterator for stepping the chain
func (chain *FondChain) Iterator() *ChainIterator {
	return &ChainIterator{
		Db:     chain.Db,
		Cursor: chain.TailHash,
	}
}

//	return true while the iterator is at the middle block
//	when iterator destinates the genesis it's false
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

	if bytes.Equal(block.Data, genesis) {
		iter.Blocker = true
		return true
	}

	if !bytes.Equal(block.PrevHash, []byte{}) {
		iter.Cursor = block.PrevHash
		return true
	}

	return false
}
