package chaincore

//		[]byte --> ?
var (
	genesis = []byte("genesis")
)

//	main entitie of block is part of chain 
type FondBlock struct {
	//		data --> tx
	Data []byte 
	
	//	hash of this 
	Hash []byte 

	PrevHash []byte 

	//	counter for pow functional
	Nonce int
}

//	produce new block before the linking with chain
func ProduceBlock(d, p []byte) FondBlock {
	block := FondBlock{
		Data: d,
		PrevHash: p,
	}

	//	init temp 
	pow := NewPow(block)
	h, n := pow.Feel()

	block.Hash = h 
	block.Nonce = n

	return block
}


//	main etitie of chain 
type FondChain struct {
	Chain []FondBlock
} 


//	init the block with data
//	link new block with fondChain 
func (chain *FondChain) LinkBlock(d []byte) {
	var idx = len(chain.Chain) - 1

	prev := chain.Chain[idx].Hash
	block := ProduceBlock(d, prev)

	chain.Chain = append(chain.Chain, block)
}


//	start fondchain with genesis block
func StartChain() *FondChain {
	return &FondChain{
		Chain: []FondBlock{ProduceBlock(genesis, []byte{})},
	}
}
