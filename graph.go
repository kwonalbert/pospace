package pospace

import (
	//"fmt"
	"os"
	//"runtime/pprof"
)

const nodeSize = hashSize

type Graph interface {
	NewNodeById(id int64, hash []byte)
	NewNode(id int64, hash []byte)
	GetId(id int64) Node
	GetNode(node int64) Node
	WriteId(node Node, id int64)
	WriteNode(node Node, id int64)
	GetParents(node, index int64) []int64
	Close()
}

type Node interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	GetHash() []byte
}

// Generate a new PoS graph of index
// Currently only supports the weaker PoS graph
// Note that this graph will have O(2^index) nodes
func NewGraph(t int, index, size, pow2, log2 int64, fn string, pk []byte) Graph {
	var db *os.File
	_, err := os.Stat(fn)
	fileExists := err == nil
	if fileExists { //file exists
		db, err = os.OpenFile(fn, os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
	} else {
		db, err = os.Create(fn)
		if err != nil {
			panic(err)
		}
	}

	g := &XiGraph{
		pk:    pk,
		fn:    fn,
		db:    db,
		index: index,
		log2:  log2,
		size:  size,
		pow2:  pow2,
	}

	if !fileExists {
		g.XiGraphIter(index)
	}

	return g
}

func subtree(log2, node int64) int64 {
	level := (log2 + 1) - Log2(node)
	return int64((1 << uint64(level)) - 1)
}

// post-order is better for disk than bfs
func bfsToPost(pow2, log2, node int64) int64 {
	if node == 0 {
		return 0
	}
	cur := node
	res := int64(0)
	for cur != 1 {
		if cur%2 == 0 {
			res -= (subtree(log2, cur) + 1)
		} else {
			res--
		}
		cur /= 2
	}
	res += 2*pow2 - 1
	return res
}
