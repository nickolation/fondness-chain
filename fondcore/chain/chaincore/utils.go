package chaincore

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"os"

	"github.com/nickolation/fondness-chain/fondcore/chain/assets"
	"github.com/nickolation/fondness-chain/fondcore/utils"
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
	var hMatrix [][]byte

	for _, tx := range block.Txn {
		hMatrix = append(hMatrix, tx.ToByte())
	}

	//	hash by tree
	tree := GrownMerkleTree(hMatrix)
	
	//	root is unic hash data of txn in this block 
	return tree.RootNode.Data
}

//	Tx to byte serialization
func (tx *Tx) ToByte() []byte {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)
	Handle(
		"tx to byte",
		enc.Encode(tx),
	)

	return buf.Bytes()
}


//	next logic 
func ToTX(buff []byte) Tx {
	return Tx{}
}

//	Bool status of existence db
func DbExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}

//	Hash tx without the unic Hash
func (tx *Tx) ToHash() []byte {
	var hash [32]byte

	t := *tx
	t.Hash = []byte{}

	hash = sha256.Sum256(t.ToByte())
	return hash[:]
}

//	Delete the sign and pubKey from tx inputs
func (tx *Tx) UnsignedTx() Tx {
	var (
		in  []InTx
		out []OutTx
	)

	for _, i := range tx.In {
		in = append(in, InTx{
			Ref:    i.Ref,
			RefIdx: i.RefIdx,
			Sign:   nil,
			PubKey: nil,
		})
	}

	for _, o := range tx.Out {
		out = append(out, OutTx{
			Force:   o.Force,
			PubHash: o.PubHash,
		})
	}

	return Tx{
		Hash: tx.Hash,
		In:   in,
		Out:  out,
	}
}

//	Validate inTx on correct sign
func (in *InTx) KeyValid(hash []byte) bool {
	lock := assets.PubHash(in.PubKey)

	return bytes.Equal(lock, hash)
}

//	Set pubKeyHash by addr decoding the base58 decoder
func (out *OutTx) KeyLock(addr []byte) {
	pHash := assets.Decode58(addr)
	out.PubHash = pHash[1 : len(pHash)-4]
}

//	Validates the pubKeyHash is correct
func (out *OutTx) IsLocked(hash []byte) bool {
	return bytes.Equal(out.PubHash, hash)
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

//	Serialize txoset to the byte
func (xoset TXOSet) ToByte() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	utils.FHandle(
		"xoset to byte",
		enc.Encode(xoset),
	)

	return buff.Bytes()
}

//	Derialize the byte to the txoset ob
func ToTXOSet(ser []byte) TXOSet {
	var xoset TXOSet

	dec := gob.NewDecoder(bytes.NewReader(ser))
	utils.FHandle(
		"xoset to byte",
		dec.Decode(&xoset),
	)

	return xoset
}





