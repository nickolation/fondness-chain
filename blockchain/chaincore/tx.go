package chaincore

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/nickolation/fondness-chain/blockchain/assets"
	"github.com/nickolation/fondness-chain/core/utils"
)

var (
	errForce    = errors.New("force isn't enouth")
	errEmptyTxn = errors.New("nil txn list")
)

type Tx struct {
	//		unic hash of tx
	Hash []byte

	In  []InTx
	Out []OutTx
}

//	Reference to the output tx inside the previous Tx
type InTx struct {
	//		transation wich is referenced by this
	Ref []byte

	//		idx of refed tx
	RefIdx int

	//		signature for the force crypro validataion
	Sign []byte

	PubKey []byte
}

//	Contains info about the loving and fondness Tx
type OutTx struct {
	//		force of the fondness
	Force int

	//		public key used to the validate and auth the loving
	PubHash []byte
}

//	Generate the new txo by addr
func ProduceTXO(val int, addr string) *OutTx {
	tx := &OutTx{
		Force:   val,
		PubHash: nil,
	}

	tx.KeyLock([]byte(addr))

	return tx
}

//	Generate new coinbase tx
//	Contains the only inIx wich isn't referenced to another out
func CoinbaseTx(addr, info string) *Tx {
	buf := make([]byte, 20)
	if info == "" {
		_, err := rand.Read(buf)
		utils.FHandle(
			"rand data generation for genesis",
			err,
		)
		//	??
		log.Printf("Fondness to - [%s]", addr)
	}

	data := fmt.Sprintf("[%x]", buf)

	in := InTx{
		Ref:    []byte{},
		RefIdx: -1,
		Sign:   nil,
		PubKey: []byte(data),
	}

	//	nil!
	out := *ProduceTXO(
		stdForce,
		addr,
	)

	tx := &Tx{
		In:  []InTx{in},
		Out: []OutTx{out},
	}
	tx.Hash = tx.ToHash()

	return tx
}

//	Makes new tx from to with amount-level
func (uxoset *UTXOSet) ProduceTx(from, to string, level int) (*Tx, error) {
	var (
		in  []InTx
		out []OutTx
	)

	mem, err := assets.AccesMemory()
	Handle(
		"accec to the memory - produce tx is locked",
		err,
	)

	h := mem.GetHeart(from)

	pubKey := h.PubKey.Key
	privKey := h.PrivKey.Key

	pubHash := assets.PubHash(pubKey)
	sum, freeOut := uxoset.FondableUTXO(pubHash, level)

	if ForceLess(sum, level) {
		Handle(
			"Balance of node isn't enouth for perf tx",
			errForce,
		)

		return nil, errForce
	}

	//	Generate sending out
	out = append(out, *ProduceTXO(
		level,
		to,
	))

	//	Generate balance/change out
	if ForceGreat(sum, level) {
		out = append(out, *ProduceTXO(
			sum-level,
			from,
		))
	}

	for sHash, outs := range freeOut {
		bHash, err := hex.DecodeString(sHash)
		Handle(
			"decode string hash to byte hash",
			err,
		)

		for _, out := range outs {
			in = append(in, InTx{
				Ref:    bHash,
				RefIdx: out,
				Sign:   nil,
				PubKey: pubKey,
			})
		}
	}

	tx := Tx{
		In:  in,
		Out: out,
	}
	tx.Hash = tx.ToHash()
	uxoset.Chain.SignTX(&tx, privKey)

	return &tx, nil
}

//	Find tx by this hash. If isn't exist it returns empty Tx and err.
func (chain *FondChain) FindTX(hash []byte) (Tx, error) {
	iter := chain.Iterator()

	for iter.Step() {
		txn := iter.Txn()

		for _, tx := range txn {
			if bytes.Equal(tx.Hash, hash) {
				return tx, nil
			}
		}
	}

	return Tx{}, errEmptyTxn
}

//	Chain wrapper under the tx sign
func (chain *FondChain) SignTX(tx *Tx, privKey ecdsa.PrivateKey) {
	mapTXs := make(map[string]Tx)

	for _, in := range tx.In {
		refTx, err := chain.FindTX(in.Ref)
		Handle(
			"ref tx search",
			err,
		)

		mapTXs[hex.EncodeToString(refTx.Hash)] = refTx
	}

	tx.Sign(privKey, mapTXs)
}

//	Chain wrapper under the tx verify
func (chain *FondChain) VefifyTX(tx *Tx) bool {
	//	cbs is simply verified
	if tx.IsCoinbase() {
		return true
	}

	mapTXs := make(map[string]Tx)

	for _, in := range tx.In {
		refTx, err := chain.FindTX(in.Ref)
		Handle(
			"ref tx search",
			err,
		)

		mapTXs[hex.EncodeToString(refTx.Hash)] = refTx
	}

	return tx.Verify(mapTXs)
}
