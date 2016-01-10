package posgraph

import (
	//"encoding/binary"
	"flag"
	//"fmt"
	//"github.com/boltdb/bolt"
	"github.com/kwonalbert/pospace/util"
	"log"
	"os"
	"testing"
	"time"
)

//exp* gets setup in test.go
var index int64 = 3
var size int64 = 0
var graphDir string = ""
var log2 int64
var pow2 int64

func TestXi(t *testing.T) {
	now := time.Now()
	_ = NewGraph(XI, graphDir, index)
	log.Printf("%d. Graph gen: %fs\n", index, time.Since(now).Seconds())

	// graph.GetDB().db.View(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte("Parents"))
	// 	c := b.Cursor()

	// 	for k, v := c.First(); k != nil; k, v = c.Next() {
	// 		key, _ := binary.Varint(k)
	// 		parents := make([]int64, len(v)/8)
	// 		for i := range parents {
	// 			parents[i], _ = binary.Varint(v[i*8 : (i+1)*8])
	// 		}
	// 	}

	// 	return nil
	// })
}

func TestEGS(t *testing.T) {
	now := time.Now()
	_ = NewGraph(EGS, graphDir, index)
	log.Printf("%d. Graph gen: %fs\n", index, time.Since(now).Seconds())

	// graph.GetDB().db.View(func(tx *bolt.Tx) error {
	// 	b := tx.Bucket([]byte("Adjlist"))
	// 	c := b.Cursor()

	// 	for k, v := c.First(); k != nil; k, v = c.Next() {
	// 		key, _ := binary.Varint(k)
	// 		adjlist := make([]int64, len(v)/8)
	// 		for i := range adjlist {
	// 			adjlist[i], _ = binary.Varint(v[i*8 : (i+1)*8])
	// 		}
	// 		log.Println(key, ":", adjlist)
	// 	}

	// 	return nil
	// })
}

func TestMain(m *testing.M) {
	size = numXi(index)
	log2 = util.Log2(size) + 1
	pow2 = int64(1 << uint64(log2))
	if (1 << uint64(log2-1)) == size {
		log2--
		pow2 = 1 << uint64(log2)
	}

	id := flag.Int("index", 1, "graph index")
	flag.Parse()
	index = int64(*id)

	graphDir = "./test"
	os.Exit(m.Run())
}
