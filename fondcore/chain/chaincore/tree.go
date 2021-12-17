package chaincore

import (
	"crypto/sha256"
	"errors"

	"github.com/nickolation/fondness-chain/fondcore/utils"
)

var (
	errNilNodes = errors.New("nil nodes list - growning is blocked")
)

//	Tree of the txn hashes.
//	Tree is essentialy the binary tree.
//	Parent of sons [1] - [2] is  sha256(sha256(1) + sha256(2))
//	The root has hight with value is log2(txn_len) and provides the unic block data
type MerkleTree struct {
	RootNode *MerkleNode
}

//	Btree node recursively
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

//	Provides the node before the hooking to the tree
func HookMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		sonsHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(sonsHashes)
		node.Data = hash[:]
	}

	node.Left = left
	node.Right = right

	return &node
}

//	Generates hashes tree by hash txn matrix.
func GrownMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	for _, d := range data {
		node := HookMerkleNode(nil, nil, d)
		nodes = append(nodes, *node)
	}

	if len(nodes) == 0 {
		utils.FHandle(
			"TREE",
			errNilNodes,
		)
	}

	for len(nodes) > 1 {
		if len(nodes)%2 != 0 {
			nodes = append(nodes, nodes[len(nodes)-1])
		}

		var level []MerkleNode
		for i := 0; i < len(nodes); i += 2 {
			node := HookMerkleNode(&nodes[i], &nodes[i+1], nil)
			level = append(level, *node)
		}

		nodes = level
	}

	//	Grown tree with root node
	tree := MerkleTree{&nodes[0]}

	return &tree
}




