package chaincore

import (
	"encoding/hex"
	"errors"
	"log"
)


var (
	errForce = errors.New("force isn't enouth")
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
	Sign string
}


//	Contains info about the loving and fondness Tx
type OutTx struct {
	//		force of the fondness
	Force int

	//		public key used to the validate and auth the loving
	PubKey string
}


//	Generate new coinbase tx 
//	Contains the only inIx wich isn't referenced to another out
func CoinbaseTx(addr, data string) *Tx {
	if data == "" {
		log.Printf("Fondness to - [%s]", addr)
	}

	in := InTx{
		Ref: []byte{},
		RefIdx: -1,
		Sign: data,
	}

	out := OutTx{
		Force: stdForce,
		PubKey: addr,
	}

	tx := &Tx{
		In: []InTx{in},
		Out: []OutTx{out},
	}
	tx.SetHash()

	return tx
}


//	Makes new tx from to with amount-level 
func (chain *FondChain) ProduceTx(from, to string, level int) (*Tx, error) {
	var (
		in []InTx
		out []OutTx
	)

	sum, freeOut := chain.FondUTXO(from, level)

	if ForceLess(sum, level) {	
		Handle(
			"Balance of node isn't enouth for perf tx",
			errForce,
		)

		return nil, errForce
	}

	//	Generate sending out
	out = append(out, OutTx{
		Force: level,
		PubKey: to,
	})

	//	Generate balance/change out
	if ForceGreat(sum, level) {
		out = append(out, OutTx{
			Force: sum - level,
			PubKey: from,
		})
	}

	for sHash, outs := range freeOut {
		bHash, err := hex.DecodeString(sHash)
		Handle(
			"decode string hash to byte hash",
			err,
		)

		for _, out := range outs {
			in = append(in, InTx{
				Ref: bHash,
				RefIdx: out,
				Sign: from,
			})
		}
	}

	tx := Tx{
		In: in,
		Out: out,
	}
	tx.SetHash()

	return &tx, nil
}
