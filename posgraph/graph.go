package posgraph

import (
	//"fmt"
	"os"
	//"runtime/pprof"
)

const hashSize = 256 / 8
const nodeSize = hashSize

const (
	TYPE1 = iota // Xi graph
	TYPE2 = iota // EGS graph
)

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
