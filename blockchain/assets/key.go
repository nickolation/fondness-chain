package assets 

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"

	"github.com/nickolation/fondness-chain/core/utils"
)


//	Wrap of the private key with elliptic key is inside 
type PrivateKey struct {
	Key ecdsa.PrivateKey 
}


//	Wrap of the public key value needs for the getting addres value in txns
type PublicKey struct {
	Key []byte
}


//	Produces key pairs with randoms unic values
func NewKeys() (PrivateKey, PublicKey) {
	curve := elliptic.P256()

	priv, err := ecdsa.GenerateKey(curve, rand.Reader)
	utils.Handle(
		"private key generation",
		err,
	)

	pub := append(priv.PublicKey.X.Bytes(), priv.Y.Bytes()...)

	pr := PrivateKey{
		Key: *priv,
	}

	pb := PublicKey{
		Key: pub,
	}

	return pr, pb
}

