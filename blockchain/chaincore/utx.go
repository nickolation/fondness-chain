package chaincore

import (
	"bytes"
	"encoding/hex"

	"github.com/dgraph-io/badger"
	"github.com/nickolation/fondness-chain/blockchain/assets"
	"github.com/nickolation/fondness-chain/core/utils"
)

var (
	utxoPrefix = []byte("utxo-")
	utxosetSize = 50000
)

type UTXOSet struct {
	Chain *FondChain
}

type TXOSet struct {
	Outs []OutTx
}

//	Delete all prefixed keys in last setSize objects.
//	Uses custom host iterator.
func (uxoset *UTXOSet) DelPrifixed(prefix []byte) {
	deleter := func(keyMatrix [][]byte) error {
		if err := uxoset.Chain.Db.Update(func(txn *badger.Txn) error {
			for _, key := range keyMatrix {
				if err := txn.Delete(key); err != nil {
					utils.Log(
						"badger delete key",
						err,
					)
					return err
				}
			}
			return nil
		}); err != nil {
			utils.FLog(
				"deleter",
				err,
			)
			return err
		}
		return nil
	}

	if err := uxoset.Chain.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		iter := txn.NewIterator(opts)
		defer iter.Close()

		host := KeysIter()

		for iter.Seek(prefix); iter.ValidForPrefix(prefix); iter.Next() {
			key := iter.Item().KeyCopy(nil)
			host.Push(key)

			if host.IsFinished() {
				if err := deleter(host.Keys); err != nil {
					return err
				}
				host.Vanish()
			}
		}

		if host.IsFree() {
			if err := deleter(host.Keys); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		utils.FLog(
			"view badger for deleting",
			err,
		)
	}
}

//	Find the uspenst tx - utx for this addr.
//	Utx contains the utxo.
func (chain *FondChain) UTX(pubHash []byte) []Tx {
	//	unspent tx
	var utx []Tx

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

				//	crypto-validation out
				if out.IsLocked(pubHash) {
					utx = append(utx, tx)
				}
			}

			//	search spentable tx
			//	coinbase isn't valid for stx map appending
			if !tx.IsCoinbase() {
				for _, in := range tx.In {
					//	crypto-validation in
					if in.KeyValid(pubHash) {
						rTxHash := hex.EncodeToString(in.Ref)
						stx[rTxHash] = append(stx[rTxHash], in.RefIdx)
					}
				}
			}
		}
	}

	return utx
}

//	Find all utxo for this addr by utxoSet
func (uxoset UTXOSet) UTXO(pubHash []byte) []OutTx {
	var utxo []OutTx

	db := uxoset.Chain.Db

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Seek(utxoPrefix); iter.ValidForPrefix(utxoPrefix); iter.Next() {
			i := iter.Item()
			v := []byte{}

			if err := i.Value(func(val []byte) error {
				v = val
				return nil
			}); err != nil {
				utils.Log(
					"value getting from utxo",
					err,
				)
				return err
			}
			outs := ToTXOSet(v)

			for _, out := range outs.Outs {
				if out.IsLocked(pubHash) {
					utxo = append(utxo, out)
				}
			}
		}
		return nil
	})

	if err != nil {
		utils.FLog(
			"find utxo from prefixed value",
			err,
		)
	}

	return utxo
}

//	Delete last prefixed utxo-value.
//	Generate the new indexed utxo set and put in the base.
func (uxoset UTXOSet) Index() {
	db := uxoset.Chain.Db

	uxoset.DelPrifixed(utxoPrefix)

	utxo := uxoset.Chain.ViewUTXO()

	err := db.Update(func(txn *badger.Txn) error {
		for txHash, outs := range utxo {
			key, err := hex.DecodeString(txHash)
			utils.Handle(
				"decode string key to the byte",
				err,
			)

			//	prefixed key
			key = append(utxoPrefix, key...)
			utils.Handle(
				"set utxo key-value",
				txn.Set(key, outs.ToByte()),
			)
		}

		return nil
	})

	utils.FHandle(
		"update db to utxo-setting",
		err,
	)
}

//	Update utxo-set in the base by the giving block txn.
func (uxoset UTXOSet) Refresh(block *FondBlock) {
	db := uxoset.Chain.Db
	outSet := TXOSet{}

	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Txn {
			if !tx.IsCoinbase() {
				for _, in := range tx.In {
					key := append(utxoPrefix, in.Ref...)

					it, err := txn.Get(key)
					utils.Handle(
						"get the utxo-set by key in",
						err,
					)

					set := []byte{}
					err = it.Value(func(val []byte) error {
						set = val
						return nil
					})
					if err != nil {
						return err
					}

					txoSet := ToTXOSet(set)

					for id, out := range txoSet.Outs {
						if id != in.RefIdx {
							outSet.Outs = append(outSet.Outs, out)
						}
					}

					if len(outSet.Outs) == 0 {
						if err = txn.Delete(key); err != nil {
							utils.Log(
								"delete utxo by key",
								err,
							)
							return err
						}
					}

					if err = txn.Set(key, outSet.ToByte()); err != nil {
						utils.Log(
							"set utxo by key",
							err,
						)
						return err
					}
				}
			}

			//	Coinbase utxo setting logic
			outSet.Outs = append(outSet.Outs, tx.Out...)
			cbsKey := append(utxoPrefix, tx.Hash...)
			if err := txn.Set(cbsKey, outSet.ToByte()); err != nil {
				utils.Log(
					"set coinbase utxo to set",
					err,
				)
				return err
			}
		}

		return nil
	})

	utils.FHandle(
		"update the utxo set by new block",
		err,
	)
}

//	Calculater the number of the tx with inner utxo.
func (uxoset UTXOSet) CountUTX() int {
	db := uxoset.Chain.Db

	ctr := 0

	db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Seek(utxoPrefix); iter.ValidForPrefix(utxoPrefix); iter.Next() {
			ctr++
		}

		return nil
	})

	return ctr
}

//	Find enough accumulated balance by this addr.
//	Is need for the performing the tx sending Node -> Node [level fondness].
//	Searching is in the utxo-set pushed at the base.
func (uxoset UTXOSet) FondableUTXO(pubHash []byte, level int) (int, map[string][]int) {
	utxoMap := make(map[string][]int)

	var sum int

	db := uxoset.Chain.Db

	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		iter := txn.NewIterator(opts)
		defer iter.Close()

		for iter.Seek(utxoPrefix); iter.ValidForPrefix(utxoPrefix); iter.Next() {
			it := iter.Item()
			key := it.Key()

			setByte := []byte{}

			it.Value(func(val []byte) error {
				setByte = val
				return nil
			})

			key = bytes.TrimPrefix(key, utxoPrefix)
			txHash := hex.EncodeToString(key)

			setOut := ToTXOSet(setByte)

			for idx, out := range setOut.Outs {
				if out.IsLocked(pubHash) && sum < level {
					sum += out.Force
					utxoMap[txHash] = append(utxoMap[txHash], idx)
				}
			}
		}

		return nil
	})

	utils.FHandle(
		"found utxo by hash and calculate need sum",
		err,
	)

	return sum, utxoMap
}

//	Get balance of Node by this addr
func (uxoset UTXOSet) GetFondness(addr string) int {
	var sum int

	pubHash := assets.Decode58([]byte(addr))
	key := pubHash[1 : len(pubHash)-4]

	utxo := uxoset.UTXO(key)
	for _, out := range utxo {
		sum += out.Force
	}

	return sum
}
