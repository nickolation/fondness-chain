package chaincore

import (
	"encoding/hex"

	"github.com/nickolation/fondness-chain/blockchain/assets"
)

//	Find the uspenst tx - utx for this addr.
//	Utx contains the utxo
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


//	Find all utxo for this addr
func (chain *FondChain) UTXO(pubHash []byte) []OutTx {
	var utxo []OutTx 

	utx := chain.UTX(pubHash)

	for _, tx := range utx {
		for _, out := range tx.Out {
			if out.IsLocked(pubHash) {
				utxo = append(utxo, out)
			}
		}
	}

	return utxo
}


//	Find enough accumulated balance by this addr.
//	Is need for the performing the tx sending Node -> Node [level fondness]
func (chain *FondChain) FondUTXO(pubHash []byte, level int) (int, map[string][]int) {
	utxoMap := make(map[string][]int)
	utx := chain.UTX(pubHash)

	var sum int 


Balance:
	for _, tx := range utx {
		txHash := hex.EncodeToString(tx.Hash) 

		for outIdx, out := range tx.Out {
			if out.IsLocked(pubHash) && ForceLess(sum, level) {
				utxoMap[txHash] = append(utxoMap[txHash], outIdx)

				sum += out.Force
				if !ForceLess(sum, level) {
					break Balance
				}
			}
		}
	}

	return sum, utxoMap
}


//	Get balance of Node by this addr
func (chain *FondChain) GetFondness(addr string) int {
	var sum int 

	pubHash := assets.Decode58([]byte(addr))
	key := pubHash[1:len(pubHash)- 4]

	utxo := chain.UTXO(key)
	for _, out := range utxo {
		sum += out.Force
	}

	return sum
}