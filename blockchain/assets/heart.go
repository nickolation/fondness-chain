package assets

var (
	version = byte(0x01)
	checksumLen = 4
)

type Heart struct {
	PrivKey PrivateKey
	PubKey PublicKey
}


//	Calculate the addr for this waller pub key
func (hrt *Heart) Addr() []byte {
	pHash := PubHash(hrt.PubKey.Key)
	vHash := append([]byte{version}, pHash...)

	chSum := Checksum(vHash)
	
	ad := append(vHash, chSum...)
	return Encode58(ad)
}


//	Heart generator with random pr, pb values 
func FeelHeart() *Heart {
	pr, pb := NewKeys()
	return &Heart{
		PrivKey: pr,
		PubKey: pb,
	}
}
