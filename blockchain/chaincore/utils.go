package chaincore

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"os"

	"github.com/nickolation/fondness-chain/core/utils"
)

const (
	//	file == existence blockchain
	dbPath = "./source/chain/MANIFEST"
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


//	Hashes all txn in the block
//	Determines pow logic part of the txn durability 
func (block *FondBlock) HashTxn() []byte {
	var (
		buff [32]byte 
		hMatrix [][]byte
	)

	for _, tx := range block.Txn {
		hMatrix = append(hMatrix, tx.Hash)
	}

	buff = sha256.Sum256(bytes.Join(hMatrix, []byte{}))
	return buff[:]
}


//	Bool status of existence db
func DbExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}


//	Hash tx
func (tx *Tx) SetHash() {
	var hash [32]byte 
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	utils.Handle(
		"encoding tx to bytes",
		enc.Encode(tx),
	)

	hash = sha256.Sum256(buff.Bytes())
	tx.Hash = hash[:]
}


//	Validate inTx on correct sign
func (in *InTx) InValid(data string) bool {
	return data == in.Sign
}


//	Validate outTx on correct pub-key
func (out *OutTx) OutValid(data string) bool {
	return data == out.PubKey
}


//	Check if the tx is coinbase 
func (tx *Tx) IsCoinbase() bool {
	i := tx.In[0]
	return len(tx.In) == 1 && len(i.Ref) == 0 && i.RefIdx == -1 
}


//	Check is x < v
func ForceLess(x, v int) bool {
	return x < v
}


//	Check is x > v
func ForceGreat(x, v int) bool {
	return x > v
}