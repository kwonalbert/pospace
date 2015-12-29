package posgraph

import (
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	//"runtime/pprof"
)

const hashSize = 256 / 8
const nodeSize = hashSize

const (
	XI  = iota
	EGS = iota
)

type GraphParam struct {
	fn string
	db *bolt.DB

	index int64
	log2  int64
	pow2  int64
	size  int64
}

type Graph interface {
	NewNode(id int64, parents []int64)
	GetParents(node, index int64) []int64
	GetSize() int64
	GetDB() *bolt.DB
	Close()
}

// Generate a new PoS graph of index
// Currently only supports the weaker PoS graph
// Note that this graph will have O(2^index) nodes
func NewGraph(t int, dir string, index int64) Graph {
	var fn string
	if t == XI {
		fn = fmt.Sprintf("%s/XI-%d", dir, index)
	} else {
		fn = fmt.Sprintf("%s/EGS-%d", dir, index)
	}

	_, err := os.Stat(fn)
	fileExists := err == nil

	db, err := bolt.Open(fn, 0600, nil)
	if err != nil {
		panic("Failed to open database")
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("Graph"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return NewXiGraph(t, !fileExists, index, db)
}
