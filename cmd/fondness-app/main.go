package main

import (
	"log"

	chain "github.com/nickolation/fondness-chain/blockchain/chaincore"
	//"github.com/nickolation/fondness-chain/cli"
	//"github.com/nickolation/fondness-chain/core/space"
)

func main() {

	bch := chain.StartChain()

	bch.LinkBlock([]byte("1"))

	for _, b := range bch.Chain {
		log.Printf("Data - [%x]", b.Data)
		log.Printf("Prev - [%x]", b.PrevHash)
		log.Printf("Hash - [%x]", b.Hash)
	}
}
