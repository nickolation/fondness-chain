package assets

import (
	"crypto/sha256"

	"github.com/mr-tron/base58/base58"
	"github.com/nickolation/fondness-chain/core/utils"
	
	//	Chnge to new 
	"golang.org/x/crypto/ripemd160"
)

//	Hashes v []byte to 58 base system for pleasure reading
func Encode58(v []byte) []byte {
	enc := base58.Encode(v)
	return []byte(enc)
}


//	Decode from 58 to the simple []byte bytes 
func Decode58(v []byte) []byte {
	dec, err :=  base58.Decode(string(v))
	utils.Handle(
		"decode to 58",
		err,
	)

	return dec
}


//	Hashe pubKey to sha256 and ripemd160
func PubHash(key []byte) []byte {
	hash256 := sha256.Sum256(key)

	hasher := ripemd160.New() 
	_, err := hasher.Write(hash256[:])
	utils.Handle(
		"ripemd160 hasher",
		err,
	)

	return hasher.Sum(nil)
}
 

//	Twice hashing of rmd.
//	Generate 4 byte checksum.
func Checksum(rmd []byte) []byte {
	fHash := sha256.Sum256(rmd)
	sHash := sha256.Sum256(fHash[:]) 

	return sHash[:checksumLen]
}














