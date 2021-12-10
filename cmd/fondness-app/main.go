package main

import (
	"log"
	"os"

	"github.com/dgraph-io/badger"
	chain "github.com/nickolation/fondness-chain/blockchain/chaincore"
	"github.com/nickolation/fondness-chain/cli"
)

func main() {
	defer os.Exit(0)

	bch := chain.StartChain()
	log.Printf("chain is - [%x]", bch.TailHash)
	defer bch.Db.Close()

	//		test-logic 
	err := bch.Db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(chain.Tail)
		if err != nil {
			log.Println(err)
		}

		err = it.Value(func(val []byte) error {
			log.Printf("MAIN: genesis [tail] is - [%x]", val)
			return nil
		})

		return err
	})

	if err != nil {
		log.Println(err)
	}

	err = bch.Db.View(func(txn *badger.Txn) error {
		it, err := txn.Get(chain.Tail)
		if err != nil {
			log.Println(err)
		}

		err = it.Value(func(val []byte) error {
			log.Printf("MAIN: block [value] is - [%x]", val)
			return nil
		})

		return err
	})

	if err != nil {
		log.Println(err)
	}

	err = cli.InitCli(bch)
	if err != nil {
		log.Printf("cli err - [%v]", err)
	}
}
