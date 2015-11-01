package pos

import (
	"encoding/json"
	"fmt"
	"github.com/steveyen/gkvlite"
	//"runtime/pprof"
	"os"
)

var graphBase string = "%s.%s%d-%d"
var nodeBase string = "%s.%d-%d"

const (
	SO = 0
	SI = 1
)

type Graph struct {
	s *gkvlite.Store
	c *gkvlite.Collection
}

type Node struct {
	Id      int      // node id
	Hash    []byte   // hash at the file
	Parents []string // parent node files
}

func (n *Node) MarshalBinary() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Node) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, n)
}

// Generate a new PoS graph of index
// Currently only supports the weaker PoS graph
// Note that this graph will have O(2^index) nodes
func NewGraph(index int, name, fn string) *Graph {
	// cpuprofile := "cpu.prof"
	// f, _ := os.Create(cpuprofile)
	// pprof.StartCPUProfile(f)
	// defer pprof.StopCPUProfile()
	// recursively generate graphs
	count := 0

	f, err := os.Create(fn)

	s, err := gkvlite.NewStore(f)
	if err != nil {
		panic(err)
	}
	c := s.SetCollection("nodes", nil)

	g := &Graph{
		s: s,
		c: c,
	}

	g.XiGraph(index, 0, name, &count)

	f.Sync()

	return g
}

// Gets the node, and update the node.
// Otherwise, create a node
func (g *Graph) GetNode(nodeName string, id int, hash []byte, parents []string) *Node {
	node := new(Node)
	val, err := g.c.Get([]byte(nodeName))
	if err != nil {
		panic(err)
	}
	if val != nil {
		node.UnmarshalBinary(val)
	} else {
		node.Id = id
		node.Hash = hash
	}
	node.Parents = append(parents, node.Parents...)
	return node
}

func (g *Graph) Write(n *Node, nodeName string) {
	b, err := n.MarshalBinary()
	if err != nil {
		panic(err)
	}
	g.c.Set([]byte(nodeName), b)
}

func numXi(index int) int {
	return (1 << uint(index)) * (index + 1) * index
}

func numButterfly(index int) int {
	return 2 * (1 << uint(index)) * index
}

// Maps a node index (0 to O(2^N)) to a folder (a physical node)
func IndexToNode(node, index, inst int, name string) string {
	sources := 1 << uint(index)
	firstButter := sources + numButterfly(index-1)
	firstXi := firstButter + numXi(index-1)
	secondXi := firstXi + numXi(index-1)
	secondButter := secondXi + numButterfly(index-1)
	sinks := secondButter + sources

	curGraph := fmt.Sprintf(graphBase, name, posName, index, inst)

	if node < sources {
		return fmt.Sprintf(nodeBase, curGraph, SO, node)
	} else if node >= sources && node < firstButter {
		node = node - sources
		butterfly0 := fmt.Sprintf(graphBase, curGraph, "C", index-1, 0)
		level := node / (1 << uint(index-1))
		nodeNum := node % (1 << uint(index-1))
		return fmt.Sprintf(nodeBase, butterfly0, level, nodeNum)
	} else if node >= firstButter && node < firstXi {
		node = node - firstButter
		return IndexToNode(node, index-1, 0, curGraph)
	} else if node >= firstXi && node < secondXi {
		node = node - firstXi
		return IndexToNode(node, index-1, 1, curGraph)
	} else if node >= secondXi && node < secondButter {
		node = node - secondXi
		butterfly1 := fmt.Sprintf(graphBase, curGraph, "C", index-1, 1)
		level := node / (1 << uint(index-1))
		nodeNum := node % (1 << uint(index-1))
		return fmt.Sprintf(nodeBase, butterfly1, level, nodeNum)
	} else if node >= secondButter && node < sinks {
		node = node - secondButter
		return fmt.Sprintf(nodeBase, curGraph, SI, node)
	} else {
		return ""
	}
}

func (g *Graph) ButterflyGraph(index, inst int, name, graph string, count *int) {
	curGraph := fmt.Sprintf(graphBase, graph, name, index, inst)
	numLevel := 2 * index
	for level := 0; level < numLevel; level++ {
		for i := 0; i < int(1<<uint(index)); i++ {
			// no parents at level 0
			nodeName := fmt.Sprintf(nodeBase, curGraph, level, i)
			if level == 0 {
				node := g.GetNode("", *count, nil, nil)
				g.Write(node, nodeName)
				*count++
				continue
			}
			prev := 0
			shift := index - level
			if level > numLevel/2 {
				shift = level - numLevel/2
			}
			if (i>>uint(shift))&1 == 0 {
				prev = i + (1 << uint(shift))
			} else {
				prev = i - (1 << uint(shift))
			}
			prev1 := fmt.Sprintf("%d-%d", level-1, prev)
			prev2 := fmt.Sprintf("%d-%d", level-1, i)
			parent1 := fmt.Sprintf("%s.%s", curGraph, prev1)
			parent2 := fmt.Sprintf("%s.%s", curGraph, prev2)

			parents := []string{parent1, parent2}
			node := g.GetNode("", *count, nil, parents)
			g.Write(node, nodeName)
			*count++
		}
	}

	err := g.s.Flush()
	if err != nil {
		panic(err)
	}
}

func (g *Graph) XiGraph(index, inst int, graph string, count *int) {
	if index == 1 {
		g.ButterflyGraph(index, inst, posName, graph, count)
		return
	}
	curGraph := fmt.Sprintf(graphBase, graph, posName, index, inst)

	for i := 0; i < int(1<<uint(index)); i++ {
		// "SO" for sources
		nodeName := fmt.Sprintf(nodeBase, curGraph, SO, i)
		node := g.GetNode("", *count, nil, nil)
		g.Write(node, nodeName)
		*count++
	}

	// recursively generate graphs
	g.ButterflyGraph(index-1, 0, "C", curGraph, count)
	g.XiGraph(index-1, 0, curGraph, count)
	g.XiGraph(index-1, 1, curGraph, count)
	g.ButterflyGraph(index-1, 1, "C", curGraph, count)

	for i := 0; i < int(1<<uint(index)); i++ {
		// "SI" for sinks
		nodeName := fmt.Sprintf(nodeBase, curGraph, SI, i)
		node := g.GetNode("", *count, nil, nil)
		g.Write(node, nodeName)
		*count++
	}

	offset := int(1 << uint(index-1)) //2^(index-1)

	// sources to sources of first butterfly
	butterfly0 := fmt.Sprintf(graphBase, curGraph, "C", index-1, 0)
	for i := 0; i < offset; i++ {
		nodeName := fmt.Sprintf(nodeBase, butterfly0, 0, i)
		parent0 := fmt.Sprintf(nodeBase, curGraph, SO, i)
		parent1 := fmt.Sprintf(nodeBase, curGraph, SO, i+offset)
		node := g.GetNode(nodeName, -1, nil, []string{parent0, parent1})
		g.Write(node, nodeName)
	}

	// sinks of first butterfly to sources of first xi graph
	xi0 := fmt.Sprintf(graphBase, curGraph, posName, index-1, 0)
	for i := 0; i < offset; i++ {
		nodeName := fmt.Sprintf(nodeBase, xi0, SO, i)
		// index is the last level; i.e., sinks
		parent := fmt.Sprintf(nodeBase, butterfly0, 2*(index-1)-1, i)
		node := g.GetNode(nodeName, -1, nil, []string{parent})
		g.Write(node, nodeName)
	}

	// sinks of first xi to sources of second xi
	xi1 := fmt.Sprintf(graphBase, curGraph, posName, index-1, 1)
	for i := 0; i < offset; i++ {
		nodeName := fmt.Sprintf(nodeBase, xi1, SO, i)
		parent := fmt.Sprintf(nodeBase, xi0, SI, i)
		if index-1 == 0 {
			parent = fmt.Sprintf(nodeBase, xi0, SO, i)
		}
		node := g.GetNode(nodeName, -1, nil, []string{parent})
		g.Write(node, nodeName)
	}

	// sinks of second xi to sources of second butterfly
	butterfly1 := fmt.Sprintf(graphBase, curGraph, "C", index-1, 1)
	for i := 0; i < offset; i++ {
		nodeName := fmt.Sprintf(nodeBase, butterfly1, 0, i)
		parent := fmt.Sprintf(nodeBase, xi1, SI, i)
		node := g.GetNode(nodeName, -1, nil, []string{parent})
		g.Write(node, nodeName)
	}

	// sinks of second butterfly to sinks
	for i := 0; i < offset; i++ {
		nodeName0 := fmt.Sprintf(nodeBase, curGraph, SI, i)
		nodeName1 := fmt.Sprintf(nodeBase, curGraph, SI, i+offset)
		parent := fmt.Sprintf(nodeBase, butterfly1, 2*(index-1)-1, i)
		node0 := g.GetNode(nodeName0, -1, nil, []string{parent})
		node1 := g.GetNode(nodeName1, -1, nil, []string{parent})
		g.Write(node0, nodeName0)
		g.Write(node1, nodeName1)
	}

	// sources to sinks directly
	for i := 0; i < int(1<<uint(index)); i++ {
		nodeName := fmt.Sprintf(nodeBase, curGraph, SI, i)
		parent := fmt.Sprintf(nodeBase, curGraph, SO, i)
		node := g.GetNode(nodeName, -1, nil, []string{parent})
		g.Write(node, nodeName)
	}

	err := g.s.Flush()
	if err != nil {
		panic(err)
	}
}