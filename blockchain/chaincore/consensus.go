package chaincore

import (
	"bytes"
	//"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"math/big"
)

//	difficulty of the hash calculating
//	determines the numbers of zeroes in the hash value
const (
	dif   = 10
	dif64 = int64(dif)
	difu  = uint(dif)

	//	max value of randomiser
	//max = 12000
)

type FondPow struct {
	Source *FondBlock
	Target *big.Int
}

func Pow(b *FondBlock) *FondPow {
	/*
	tg, err := rand.Int(rand.Reader, big.NewInt(max))
	Handle(
		"error with generate random bigInt",
		err,
	) */

	tg := big.NewInt(1)

	tg.Lsh(tg, 256-difu)
	log.Printf("target is  - [%x]", tg)

	return &FondPow{
		Source: b,
		Target: tg,
	}
}

//	calculate temp hash by nonce
func (pow *FondPow) CalcHash(nonce int) []byte {
	n, err := Hex(int64(nonce))
	Handle(
		"Nonce hexing error",
		err,
	)

	d, _ := Hex(dif64)

	return bytes.Join(
		[][]byte{
			pow.Source.PrevHash,
			pow.Source.HashTxn(),
			n,
			d,
		},
		[]byte{},
	)
}

//	calculate hash of block within the target frames
//	take the hash and crypto ctr
func (pow *FondPow) Feel() ([]byte, int) {
	var nonce = 0
	var bHash big.Int

	hash := [32]byte{}

	for nonce < math.MaxInt64 {
		d := pow.CalcHash(nonce)
		hash = sha256.Sum256(d)
		bHash.SetBytes(hash[:])

		if bHash.Cmp(pow.Target) == -1 {
				break
		} else {
				nonce++
		}
	}

	log.Printf("Nonce is - [%d]", nonce)
	return hash[:], nonce
}

//	validate hash on the target frame
func (pow *FondPow) Validate() bool {
	var cmp big.Int

	b := pow.CalcHash(pow.Source.Nonce)
	hash := sha256.Sum256(b)

	cmp.SetBytes(hash[:])

	return cmp.Cmp(pow.Target) == -1
}

//	convert number to hex string
func Hex(num int64) ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
