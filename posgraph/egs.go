package posgraph

import (
	//"fmt"
	"github.com/kwonalbert/pospace/util"
)

type EGSGraph struct {
	Graph_
}

// generate graph according to "On sparse graphs with dense long paths"
func NewEGSGraph(t int, gen bool, size int64, db DB) *EGSGraph {
	g := &EGSGraph{
		Graph_{
			size: size,
			t:    EGS,
		},
	}

	log2 := util.Log2(size) + 1
	pow2 := int64(1 << uint64(log2))
	if (1 << uint64(log2-1)) == size {
		log2--
		pow2 = 1 << uint64(log2)
	}

	g.pow2 = pow2
	g.log2 = log2
	g.db = db

	if gen {
		g.EGSGraph()
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

func (g *EGSGraph) EGSGraph() {
	// create 2^n-1 vertices, and edges
	// (i) from the paper
	for i := int64(1); i <= g.size; i++ {
		var parents []int64
		for j := util.Max(0, i-4*g.log2); j < i; j++ {
			parents = append(parents, j)
		}
		g.NewNodeP(i, parents)
	}

	// (ii) from the paper
	tBound := util.Log2(g.log2/2) + 1
	if (1 << uint64(tBound-1)) == (g.log2 / 2) {
		tBound--
	}

	for t := tBound; t < g.log2; t++ {
		tpow2 := int64(1 << uint64(t))
		for m := int64(0); m < int64(1<<uint64(g.log2-tBound)); m++ {
			for i := int64(1); i <= 10; i++ {
				if (m+i)*tpow2 > g.size {
					continue
				}
				//TODO: figure out what ep1 is really..
				ep1 := float64(0.88)
				srcs := g.dGraph(m*tpow2, tpow2)
				sinks := g.dGraph((m+1)*tpow2, tpow2)
				g.bipartiteGraph(srcs, sinks, ep1)
			}
		}
	}
}

// Generates a bipartite graph that satisfies Lemma 1 from the paper
// TODO: currently generates a random bipartite graph
//       should check if the generated graph satisfies the properties
func (g *EGSGraph) bipartiteGraph(srcs, sinks []int64, delta float64) {
	numEdges := int64(delta * float64(len(sinks)))

	// TODO: make this more efficient;
	//       currently reads the node everytime to update parents list..
	for _, s := range srcs {
		vals := util.NRandRange(0, int64(len(sinks)), numEdges)
		for i := range vals {
			parents := util.Union(g.GetParents(sinks[vals[i]]), []int64{s})
			g.NewNodeP(sinks[vals[i]], parents)
		}
	}
}
