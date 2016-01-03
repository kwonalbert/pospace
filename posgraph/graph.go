package posgraph

import (
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	//"runtime/pprof"
)

const (
	XI  = iota
	EGS = iota
)

type Graph_ struct {
	fn string
	db *bolt.DB

	index int64
	log2  int64
	pow2  int64
	size  int64
}

type Graph interface {
	NewNodeP(id int64, parents []int64)
	GetParents(id int64) []int64
	NewNodeA(id int64, adjlist []int64)
	GetAdjacency(id int64) []int64
	GetSize() int64
	GetDB() *bolt.DB
	ChangeDB(*bolt.DB)
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

	var db *bolt.DB
	if fileExists { //open it as read only
		db, err = bolt.Open(fn, 0600, &bolt.Options{ReadOnly: true})
	} else {
		db, err = bolt.Open(fn, 0600, nil)
	}
	if err != nil {
		panic("Failed to open database")
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("Parents"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("Adjlist"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	g := NewXiGraph(t, !fileExists, index, db)
	g.Close()
	db, err = bolt.Open(fn, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		panic("Failed to open database")
	}
	g.ChangeDB(db)

	return g
}

func (g *Graph_) NewNodeP(node int64, parents []int64) {
	// header := *(*reflect.SliceHeader)(unsafe.Pointer(&parents))
	// header.Len *= 8
	// header.Cap *= 8
	// data := *(*[]byte)(unsafe.Pointer(&header))
	// log.Println("New node:", node, parents)

	key := make([]byte, 8)
	binary.PutVarint(key, node)
	data := make([]byte, len(parents)*8)
	for i := range parents {
		binary.PutVarint(data[i*8:(i+1)*8], parents[i])
	}

	g.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Parents"))
		err := b.Put(key, data)
		return err
	})
}

func (g *Graph_) NewNodeA(id int64, adjlist []int64) {
	// header := *(*reflect.SliceHeader)(unsafe.Pointer(&parents))
	// header.Len *= 8
	// header.Cap *= 8
	// data := *(*[]byte)(unsafe.Pointer(&header))
	// log.Println("New node:", id, parents)

	key := make([]byte, 8)
	binary.PutVarint(key, id)
	data := make([]byte, len(adjlist)*8)
	for i := range adjlist {
		binary.PutVarint(data[i*8:(i+1)*8], adjlist[i])
	}

	g.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Adjlist"))
		err := b.Put(key, data)
		return err
	})
}

func (g *Graph_) GetParents(id int64) []int64 {
	if id < int64(1<<uint64(g.index)) {
		return nil
	}

	key := make([]byte, 8)
	binary.PutVarint(key, id)

	var data []byte
	g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Parents"))
		data = b.Get(key)
		return nil
	})

	parents := make([]int64, len(data)/8)
	for i := range parents {
		parents[i], _ = binary.Varint(data[i*8 : (i+1)*8])
	}

	return parents
}

func (g *Graph_) GetAdjacency(id int64) []int64 {
	if id < int64(1<<uint64(g.index)) {
		return nil
	}

	key := make([]byte, 8)
	binary.PutVarint(key, id)

	var data []byte
	g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Adjlist"))
		data = b.Get(key)
		return nil
	})

	adjlist := make([]int64, len(data)/8)
	for i := range adjlist {
		adjlist[i], _ = binary.Varint(data[i*8 : (i+1)*8])
	}

	return adjlist
}

func (g *Graph_) GetSize() int64 {
	return g.size
}

func (g *Graph_) GetDB() *bolt.DB {
	return g.db
}

func (g *Graph_) GetType() int {
	return XI
}

func (g *Graph_) ChangeDB(db *bolt.DB) {
	g.db = db
}

func (g *Graph_) Close() {
	g.db.Close()
}
