package chaincore

import "encoding/hex"
 
//	Find the uspenst tx - utx for this addr.
//	Utx contains the utxo
func (chain *FondChain) UTX(addr string) []Tx {
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
				if out.OutValid(addr) {
					utx = append(utx, tx)
				}
			}

			//	search spentable tx
			//	coinbase isn't valid for stx map appending
			if !tx.IsCoinbase() {
				for _, in := range tx.In {
					//	crypto-validation in
					if in.InValid(addr) {
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
func (chain *FondChain) UTXO(addr string) []OutTx {
	var utxo []OutTx 

	utx := chain.UTX(addr)

	for _, tx := range utx {
		for _, out := range tx.Out {
			if out.OutValid(addr) {
				utxo = append(utxo, out)
			}
		}
	}

	return utxo
}


//	Find enough accumulated balance by this addr.
//	Is need for the performing the tx sending Node -> Node [level fondness]
func (chain *FondChain) FondUTXO(addr string, level int) (int, map[string][]int) {
	utxoMap := make(map[string][]int)
	utx := chain.UTX(addr)

	var sum int 


Balance:
	for _, tx := range utx {
		txHash := hex.EncodeToString(tx.Hash) 

		for outIdx, out := range tx.Out {
			if out.OutValid(addr) && ForceLess(sum, level) {
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

	utxo := chain.UTXO(addr)
	for _, out := range utxo {
		sum += out.Force
	}

	return sum
}