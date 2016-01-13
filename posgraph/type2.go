package posgraph

import (
	"github.com/kwonalbert/pospace/util"
	//"log"
)

// "Full" proof-of-space graph
type Type2Graph struct {
	Graph_
	m int64
}

func NewType2Graph(t int, gen bool, index int64, db DB) *Type2Graph {
	indexpow2 := int64(1 << uint64(index))
	//TODO: get the correct constant here
	m := indexpow2 / index
	g := &Type2Graph{
		Graph_{
			index: index,
			size:  m * index,
			t:     TYPE2,
		},
		m,
	}

	size := g.GetSize()
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
		g.Type2Graph()
	}

	return g
}

func (g *Type2Graph) Type2Graph() {
	egs := NewGraph(EGS, "/tmp", g.index)

	for i := int64(0); i < g.index; i++ {
		parents := egs.GetParents(i)
		for _, p := range parents {
			g.bipartiteGraph(i*g.m, p*g.m)
		}
	}
}

// Generate random bipartite graph betwene srcs and sinks
func (g *Type2Graph) bipartiteGraph(src_start, sink_start int64) {
	for s := src_start; s < src_start+g.m; s++ {
		numEdges := util.Rand(g.m)
		vals := util.NRandRange(sink_start, sink_start+g.m, numEdges)
		for _, t := range vals {
			parents := util.Union(g.GetParents(t), []int64{s})
			g.NewNodeP(t, parents)
		}
	}
}
