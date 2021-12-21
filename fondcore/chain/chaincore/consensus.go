package chaincore

import (
	"bytes"
	"errors"

	//"crypto/rand"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"

	"github.com/nickolation/fondness-chain/fondcore/utils"
)

var (
	errNilTx = errors.New("nil tx in mapTXs")
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
	//log.Printf("target is  - [%x]", tg)

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

//	Sign all inTx of this tx by privKey
func (tx *Tx) Sign(privKey ecdsa.PrivateKey, mapTXs map[string]Tx) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.In {
		if mapTXs[hex.EncodeToString(in.Ref)].Hash == nil {
			utils.FLog(
				"nil tx in mapTXs",
				errNilTx,
			)
		}
	}

	unsTx := tx.UnsignedTx()

	for id, in := range unsTx.In {
		refTx := mapTXs[hex.EncodeToString(in.Ref)]

		//	???
		unsTx.In[id].Sign = nil
		unsTx.In[id].PubKey = refTx.Out[in.RefIdx].PubHash

		unsTx.Hash = unsTx.ToHash()
		unsTx.In[id].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, unsTx.Hash)
		Handle(
			"sign unsigned tx hash",
			err,
		)

		sign := append(r.Bytes(), s.Bytes()...)
		tx.In[id].Sign = sign
	}
}

//	Verify tx pubKey on pubKey and signature correction
func (tx *Tx) Verify(mapTXs map[string]Tx) bool {
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.In {
		if mapTXs[hex.EncodeToString(in.Ref)].Hash == nil {
			utils.FLog(
				"nil tx in mapTXs",
				errNilTx,
			)
		}
	}

	unsTx := tx.UnsignedTx()
	curve := elliptic.P256()

	for id, in := range tx.In {
		refTx := mapTXs[hex.EncodeToString(in.Ref)]
		unsTx.In[id].Sign = nil

		unsTx.In[id].PubKey = refTx.Out[in.RefIdx].PubHash
		unsTx.Hash = unsTx.ToHash()
		unsTx.In[id].PubKey = nil

		var (
			r = big.Int{}
			s = big.Int{}
		)

		sCoord := len(in.Sign) / 2

		r.SetBytes(in.Sign[:(sCoord)])
		s.SetBytes(in.Sign[sCoord:])

		var (
			x = big.Int{}
			y = big.Int{}
		)

		kCoord := len(in.PubKey) / 2
		x.SetBytes(in.PubKey[:kCoord])
		y.SetBytes(in.PubKey[kCoord:])

		pubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}
		if !ecdsa.Verify(&pubKey, unsTx.Hash, &r, &s) {
			return false
		}
	}

	return true
}
