package posgraph

import (
	"encoding/binary"
	"github.com/boltdb/bolt"
	"github.com/kwonalbert/pospace/util"
	//"log"

	// "reflect"
	// "unsafe"
)

type XiGraph struct {
	GraphParam
}

func NewXiGraph(t int, gen bool, index int64, db *bolt.DB) *XiGraph {
	g := &XiGraph{
		GraphParam{
			index: index,
			size:  numXi(index),
		},
	}

	size := g.GetSize()
	log2 := util.Log2(size) + 1
	pow2 := int64(1 << uint64(log2))
	if (1 << uint64(log2-1)) == size {
		log2--
		pow2 = 1 << uint64(log2)
	}

	g.size = size
	g.pow2 = pow2
	g.log2 = log2
	g.db = db

	if gen {
		g.XiGraphIter(index)
	}

	return g
}

func (g *XiGraph) NewNode(node int64, parents []int64) {
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
		b := tx.Bucket([]byte("Graph"))
		err := b.Put(key, data)
		return err
	})
}

// compute parents of nodes
func (g *XiGraph) GetParents(node, index int64) []int64 {
	if node < int64(1<<uint64(index)) {
		return nil
	}

	key := make([]byte, 8)
	binary.PutVarint(key, node)

	var data []byte
	g.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Graph"))
		data = b.Get(key)
		return nil
	})

	parents := make([]int64, len(data)/8)
	for i := range parents {
		parents[i], _ = binary.Varint(data[i*8 : (i+1)*8])
	}

	return parents
}

func (g *XiGraph) GetSize() int64 {
	return g.size
}

func (g *XiGraph) GetDB() *bolt.DB {
	return g.db
}

func (g *XiGraph) GetType() int {
	return XI
}

func (g *XiGraph) ChangeDB(db *bolt.DB) {
	g.db = db
}

func (g *XiGraph) Close() {
	g.db.Close()
}

func numXi(index int64) int64 {
	return (1 << uint64(index)) * (index + 1) * index
}

func (g *XiGraph) butterflyGraph(index int64, count *int64) {
	if index == 0 {
		index = 1
	}
	numLevel := 2 * index
	perLevel := int64(1 << uint64(index))
	begin := *count - perLevel // level 0 created outside
	// no parents at level 0
	var level, i int64
	for level = 1; level < numLevel; level++ {
		for i = 0; i < perLevel; i++ {
			var prev int64
			shift := index - level
			if level > numLevel/2 {
				shift = level - numLevel/2
			}
			if (i>>uint64(shift))&1 == 0 {
				prev = i + (1 << uint64(shift))
			} else {
				prev = i - (1 << uint64(shift))
			}
			parents := []int64{begin + (level-1)*perLevel + prev,
				*count - perLevel}

			g.NewNode(*count, parents)
			*count++
		}
	}
}

// Iterative generation of the graph
func (g *XiGraph) XiGraphIter(index int64) {
	count := int64(0)

	stack := []int64{index, index, index, index, index}
	graphStack := []int{4, 3, 2, 1, 0}

	var i int64
	graph := 0
	pow2index := int64(1 << uint64(index))
	for i = 0; i < pow2index; i++ { //sources at this level
		g.NewNode(count, nil)
		count++
	}

	if index == 1 {
		g.butterflyGraph(index, &count)
		return
	}

	for len(stack) != 0 && len(graphStack) != 0 {
		index, stack = stack[len(stack)-1], stack[:len(stack)-1]
		graph, graphStack = graphStack[len(graphStack)-1], graphStack[:len(graphStack)-1]

		indices := []int64{index - 1, index - 1, index - 1, index - 1, index - 1}
		graphs := []int{4, 3, 2, 1, 0}

		pow2index := int64(1 << uint64(index))
		pow2index_1 := int64(1 << uint64(index-1))

		if graph == 0 {
			sources := count - pow2index
			// sources to sources of first butterfly
			// create sources of first butterly
			for i = 0; i < pow2index_1; i++ {
				parents := []int64{sources + i,
					sources + i + pow2index_1}

				g.NewNode(count, parents)
				count++
			}
		} else if graph == 1 {
			firstXi := count
			// sinks of first butterfly to sources of first xi graph
			for i = 0; i < pow2index_1; i++ {
				nodeId := firstXi + i
				parents := []int64{firstXi - pow2index_1 + i}
				g.NewNode(nodeId, parents)
				count++
			}
		} else if graph == 2 {
			secondXi := count
			// sinks of first xi to sources of second xi
			for i = 0; i < pow2index_1; i++ {
				nodeId := secondXi + i
				parents := []int64{secondXi - pow2index_1 + i}
				g.NewNode(nodeId, parents)
				count++
			}
		} else if graph == 3 {
			secondButter := count
			// sinks of second xi to sources of second butterfly
			for i = 0; i < pow2index_1; i++ {
				nodeId := secondButter + i
				parents := []int64{secondButter - pow2index_1 + i}
				g.NewNode(nodeId, parents)
				count++
			}
		} else {
			sinks := count
			sources := sinks + pow2index - numXi(index)
			for i = 0; i < pow2index_1; i++ {
				nodeId0 := sinks + i
				nodeId1 := sinks + i + pow2index_1

				parents0 := []int64{sinks - pow2index_1 + i,
					sources + i}
				parents1 := []int64{sinks - pow2index_1 + i,
					sources + i + pow2index_1}

				g.NewNode(nodeId0, parents0[:])
				g.NewNode(nodeId1, parents1[:])
				count += 2
			}
		}

		if (graph == 0 || graph == 3) ||
			((graph == 1 || graph == 2) && index == 2) {
			g.butterflyGraph(index-1, &count)
		} else if graph == 1 || graph == 2 {
			stack = append(stack, indices...)
			graphStack = append(graphStack, graphs...)
		}
	}
}
