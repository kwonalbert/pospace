package posgraph

import (
	"encoding/binary"
	"fmt"
	"github.com/boltdb/bolt"
	"os"
	// "reflect"
	// "unsafe"
)

const (
	TYPE1 = iota
	EGS   = iota
	TYPE2 = iota
)

// creating another DB type so it's easier to change underlying DB later
type DB struct {
	db *bolt.DB
}

type Graph_ struct {
	fn string
	db DB

	index int64
	log2  int64
	pow2  int64
	size  int64

	t int //type of the graph
}

type Graph interface {
	NewNodeP(id int64, parents []int64)
	GetParents(id int64) []int64
	NewNodeA(id int64, adjlist []int64)
	GetAdjacency(id int64) []int64
	GetSize() int64
	GetDB() DB
	ChangeDB(DB)
	Close()
}

// Generate a new PoS graph of index
// Currently only supports the weaker PoS graph
// Note that this graph will have O(2^index) nodes
func NewGraph(t int, dir string, index int64) Graph {
	var fn string
	if t == TYPE1 {
		fn = fmt.Sprintf("%s/T1-%d", dir, index)
	} else if t == EGS {
		fn = fmt.Sprintf("%s/EGS-%d", dir, index)
	} else if t == TYPE2 {
		fn = fmt.Sprintf("%s/T2-%d", dir, index)
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

	var g Graph
	if t == TYPE1 {
		g = NewType1Graph(t, !fileExists, index, DB{db})
	} else if t == EGS {
		//'index' for EGS is overloaded to be size
		g = NewEGSGraph(t, !fileExists, index, DB{db})
	} else if t == TYPE2 {
		g = NewType2Graph(t, !fileExists, index, DB{db})
	}

	// a hack for testing; graph should be opened for read only after gen
	g.Close()
	db, err = bolt.Open(fn, 0600, &bolt.Options{ReadOnly: true})
	if err != nil {
		panic("Failed to open database")
	}
	g.ChangeDB(DB{db})

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

	g.db.db.Update(func(tx *bolt.Tx) error {
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

	g.db.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Adjlist"))
		err := b.Put(key, data)
		return err
	})
}

func (g *Graph_) GetParents(id int64) []int64 {
	// first pow2 nodes are srcs in TYPE1
	if g.t == TYPE1 && id < int64(1<<uint64(g.index)) {
		return nil
	}

	key := make([]byte, 8)
	binary.PutVarint(key, id)

	var data []byte
	g.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Parents"))
		d := b.Get(key)
		data = make([]byte, len(d))
		copy(data, d)
		return nil
	})

	parents := make([]int64, len(data)/8)
	for i := range parents {
		parents[i], _ = binary.Varint(data[i*8 : (i+1)*8])
	}

	return parents
}

func (g *Graph_) GetAdjacency(id int64) []int64 {
	key := make([]byte, 8)
	binary.PutVarint(key, id)

	var data []byte
	g.db.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Adjlist"))
		d := b.Get(key)
		data = make([]byte, len(d))
		copy(data, d)
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

func (g *Graph_) GetDB() DB {
	return g.db
}

func (g *Graph_) GetType() int {
	return g.t
}

func (g *Graph_) ChangeDB(db DB) {
	g.db = db
}

func (g *Graph_) Close() {
	g.db.db.Close()
}
