package chaincore

import (
	"fmt"
	"strings"
)

//	Tx interface of printing provides all valid data of this tx
func (tx Tx) String() string {
	var info []string

	info = append(info, fmt.Sprintf("-----TX - [%x]:", tx.Hash))

	for i, in := range tx.In {
		info = append(info, fmt.Sprintf("  Input %d:", i))
		info = append(info, fmt.Sprintf("  TX-HASH:     %x", in.Ref))
		info = append(info, fmt.Sprintf("  IDX:       %d", in.RefIdx))
		info = append(info, fmt.Sprintf("  Signature: %x", in.Sign))
		info = append(info, fmt.Sprintf("  PubKey:    %x\n\n", in.PubKey))
	}

	for i, out := range tx.Out {
		info = append(info, fmt.Sprintf("  Output %d:", i))
		info = append(info, fmt.Sprintf("  Force:  %d", out.Force))
		info = append(info, fmt.Sprintf("  PubHash: %x\n\n", out.PubHash))
	}

	return strings.Join(info, "\n")
}

//	Block interface of printing provides all valid data of this block
func (block FondBlock) String() string {
	var info []string

	info = append(info, fmt.Sprintf("-----BLOCK - [%x]:", block.Hash))
	info = append(info, fmt.Sprintf("  Prev Hash - [%x]:", block.PrevHash))
	info = append(info, fmt.Sprintf("  Nonce - [%d]:", block.Nonce))

	return strings.Join(info, "\n")
}



