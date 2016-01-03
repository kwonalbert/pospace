package posgraph

import (
	"github.com/boltdb/bolt"
	"github.com/kwonalbert/pospace/util"
	// "golang.org/x/crypto/sha3"
)

type EGSGraph struct {
	Graph_
}

// generate graph according to "On sparse graphs with dense long paths"
func NewEGSGraph(t int, gen bool, index int64, db *bolt.DB) *EGSGraph {
	return nil
}

func dGraph(index, v, m int64) []int64 {
	pow2 := int64(1 << uint64(index))
	var d []int64
	for i := v; i < util.Min(pow2, v+m-1); i++ {
		d = append(d, i)
	}
	return d
}

func (g *EGSGraph) EGSGraphIter(index int64) {
	// create 2^n-1 vertices, and edges
	// (i) from the paper
	pow2 := int64(1 << uint64(index))
	for i := int64(0); i < pow2; i++ {
		var parents []int64
		for j := util.Max(i-4*index+1, 0); j < i; j++ {
			parents = append(parents, j)
		}
		g.NewNodeP(i, parents)
	}

	// (ii) from the paper
	tBound := util.Log2(index/2) + 1
	if (1 << uint64(tBound-1)) == (index / 2) {
		tBound--
	}

	for t := tBound; t < index; t++ {
		for m := int64(0); m < int64(1<<uint64(index-tBound)); m++ {
			for i := int64(1); i <= 10; i++ {
				tpow2 := int64(1 << uint64(t))
				if (m+i)*tpow2 > pow2 {
					continue
				}
				//TODO: figure out what this is really..
				ep1 := float64(0.99)
				srcs := dGraph(index, m*tpow2, tpow2)
				sinks := dGraph(index, (m+1)*tpow2, tpow2)
				g.BipartiteGraph(srcs, sinks, ep1)
			}
		}
	}
}

// Generates a bipartite graph that satisfies Lemma 1 from the paper
// TODO: currently generates a random bipartite graph
//       should check if the generated graph satisfies the properties
func (g *EGSGraph) BipartiteGraph(srcs, sinks []int64, delta float64) {
	if len(srcs) != len(sinks) {
		panic("srcs and sinks need to be the same size!")
	}

	numEdges := int64(delta * float64(len(sinks)))

	for _, s := range srcs {
		vals := util.NRandRange(0, int64(len(sinks)), numEdges)
		var adjlist []int64
		for i := range vals {
			adjlist = append(adjlist, sinks[vals[i]])
		}
		adjlist = util.Union(g.GetAdjacency(s), adjlist)
		g.NewNodeA(s, adjlist)
	}
}
