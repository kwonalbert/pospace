package posgraph

import (
	//"fmt"
	"github.com/kwonalbert/pospace/util"
	// "golang.org/x/crypto/sha3"
)

type EGSGraph struct {
	Graph_
}

// generate graph according to "On sparse graphs with dense long paths"
func NewEGSGraph(t int, gen bool, index int64, db DB) *EGSGraph {
	g := &EGSGraph{
		Graph_{
			index: index,
			size:  int64(1 << uint64(index)),
			t:     EGS,
		},
	}

	g.pow2 = g.size
	g.log2 = index
	g.db = db

	if gen {
		g.EGSGraph(index)
	}

	return g
}

func (g *EGSGraph) dGraph(v, m int64) []int64 {
	var d []int64
	for i := v; i < util.Min(g.pow2, v+m-1); i++ {
		d = append(d, i)
	}
	return d
}

func (g *EGSGraph) EGSGraph(index int64) {
	// create 2^n-1 vertices, and edges
	// (i) from the paper
	pow2 := int64(1 << uint64(index))
	for i := int64(0); i < pow2; i++ {
		var adjlist []int64
		for j := i + 1; j < util.Min(pow2, i+4*index); j++ {
			adjlist = append(adjlist, j)
		}
		g.NewNodeA(i, adjlist)
	}

	// (ii) from the paper
	tBound := util.Log2(index/2) + 1
	if (1 << uint64(tBound-1)) == (index / 2) {
		tBound--
	}

	for t := tBound; t < index; t++ {
		tpow2 := int64(1 << uint64(t))
		for m := int64(0); m < int64(1<<uint64(index-tBound)); m++ {
			for i := int64(1); i <= 10; i++ {
				if (m+i)*tpow2 > pow2 {
					continue
				}
				//TODO: figure out what ep1 is really..
				ep1 := float64(0.88)
				srcs := g.dGraph(m*tpow2, tpow2)
				sinks := g.dGraph((m+1)*tpow2, tpow2)
				g.BipartiteGraph(srcs, sinks, ep1)
			}
		}
	}
}

// Generates a bipartite graph that satisfies Lemma 1 from the paper
// TODO: currently generates a random bipartite graph
//       should check if the generated graph satisfies the properties
func (g *EGSGraph) BipartiteGraph(srcs, sinks []int64, delta float64) {
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
