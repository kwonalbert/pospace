package prover

import (
	"encoding/binary"
	"fmt"
	"github.com/kwonalbert/pospace/posgraph"
	"github.com/kwonalbert/pospace/util"
	"golang.org/x/crypto/sha3"
	"os"
)

const hashSize = 32

type Prover struct {
	pk    []byte
	graph posgraph.Graph // storage for all the graphs

	commit []byte   // root hash of the merkle tree
	space  *os.File // file that stores all hashes

	pow2  int64 // next closest power of 2 of size
	log2  int64 // log2 of pow2
	empty map[int64]bool
}

type Commitment struct {
	Pk     []byte
	Commit []byte
}

func NewProver(pk []byte, index int64, graphDir, spaceDir string) *Prover {
	g := posgraph.NewGraph(posgraph.TYPE1, graphDir, index)

	size := g.GetSize()
	log2 := util.Log2(size) + 1
	pow2 := int64(1 << uint64(log2))
	if (1 << uint64(log2-1)) == size {
		log2--
		pow2 = 1 << uint64(log2)
	}

	empty := make(map[int64]bool)
	// if not power of 2, then uneven merkle
	// mark all the empty ones
	if util.Count(uint64(size)) != 1 {
		for i := pow2 + size; util.Count(uint64(i+1)) != 1; i /= 2 {
			empty[i+1] = true
		}
	}

	f, err := os.Create(fmt.Sprintf("%s/Space-%d", spaceDir, index))
	if err != nil {
		panic(err)
	}

	p := Prover{
		pk:    pk,
		graph: g,
		space: f,

		pow2:  pow2,
		log2:  log2,
		empty: empty,
	}
	return &p
}

func (p *Prover) GetHash(id int64) []byte {
	data := make([]byte, hashSize)
	n, err := p.space.ReadAt(data, id*hashSize)
	if err != nil || n != hashSize {
		panic(err)
	}
	return data
}

func (p *Prover) PutHash(id int64, data []byte) {
	n, err := p.space.WriteAt(data, id*hashSize)
	if err != nil || n != hashSize {
		panic(err)
	}
}

// Assuming topo-sorted..
func (p *Prover) initGraph() {
	for i := int64(0); i < p.graph.GetSize(); i++ {
		var ph []byte
		parents := p.graph.GetParents(i)
		for _, parent := range parents {
			pid := util.BfsToPost(p.pow2, p.log2, parent+p.pow2)
			ph = append(ph, p.GetHash(pid)...)
		}
		buf := make([]byte, 8)
		binary.PutVarint(buf, i)
		buf = append(p.pk, buf...)
		buf = append(buf, ph...)
		hash := sha3.Sum256(buf)
		id := util.BfsToPost(p.pow2, p.log2, i+p.pow2)
		p.PutHash(id, hash[:])
	}
}

// Generate a merkle tree of the hashes of the vertices
// return: root hash of the merkle tree
//         will also write out the merkle tree
func (p *Prover) Init() *Commitment {
	// build the merkle tree in depth first fashion
	// root node is 1
	p.initGraph()
	root := p.generateMerkle()
	p.commit = root

	commit := &Commitment{
		Pk:     p.pk,
		Commit: root,
	}

	return commit
}

// Read the commitment from pre-initialized graph
func (p *Prover) PreInit() *Commitment {
	hash := p.GetHash(2*p.pow2 - 1)
	p.commit = hash
	commit := &Commitment{
		Pk:     p.pk,
		Commit: p.commit,
	}
	return commit
}

func (p *Prover) emptyMerkle(node int64) bool {
	_, found := p.empty[node]
	return found
}

// Iterative function to generate merkle tree
// Should have at most O(lgn) hashes in memory at a time
// return: the root hash
func (p *Prover) generateMerkle() []byte {
	var stack []int64
	var hashStack [][]byte

	cur := int64(1)
	count := int64(1)

	for count == 1 || len(stack) != 0 {
		empty := p.emptyMerkle(cur)
		for ; cur < 2*p.pow2 && !empty; cur *= 2 {
			if cur < p.pow2 { //right child
				stack = append(stack, 2*cur+1)
			}
			stack = append(stack, cur)
		}

		if empty {
			count += util.Subtree(p.log2, cur)
			hashStack = append(hashStack, make([]byte, hashSize))
		}

		cur, stack = stack[len(stack)-1], stack[:len(stack)-1]

		if len(stack) > 0 && cur < p.pow2 &&
			(stack[len(stack)-1] == 2*cur+1) {
			stack = stack[:len(stack)-1]
			stack = append(stack, cur)
			cur = 2*cur + 1
			continue
		}

		if cur >= p.pow2 {
			if cur >= p.pow2+p.graph.GetSize() {
				hashStack = append(hashStack, make([]byte, hashSize))
				count++
			} else {
				hash := p.GetHash(count)
				count++
				hashStack = append(hashStack, hash)
			}
		} else if !p.emptyMerkle(cur) {
			hash2 := hashStack[len(hashStack)-1]
			hashStack = hashStack[:len(hashStack)-1]
			hash1 := hashStack[len(hashStack)-1]
			hashStack = hashStack[:len(hashStack)-1]
			val := append(hash1[:], hash2[:]...)
			hash := sha3.Sum256(val)

			hashStack = append(hashStack, hash[:])

			p.PutHash(count, hash[:])
			count++
		}
		cur = 2 * p.pow2
	}

	return hashStack[0]
}

// Open a node in the merkle tree
// return: hash of node, and the lgN hashes to verify node
func (p *Prover) Open(node int64) ([]byte, [][]byte) {
	hash := p.GetHash(util.BfsToPost(p.pow2, p.log2, node+p.pow2))

	proof := make([][]byte, p.log2)
	count := 0
	for i := node + p.pow2; i > 1; i /= 2 { // root hash not needed, so >1
		var sib int64

		if i%2 == 0 { // need to send only the sibling
			sib = i + 1
		} else {
			sib = i - 1
		}

		if sib >= p.pow2+p.graph.GetSize() || p.emptyMerkle(sib) {
			proof[count] = make([]byte, hashSize)
		} else {
			proof[count] = p.GetHash(util.BfsToPost(p.pow2, p.log2, sib))
		}
		count++
	}
	return hash, proof
}

// Receives challenges from the verifier to prove PoS
// return: the hash values of the challenges, the parent hashes,
//         the proof for each, and the proof for the parents
func (p *Prover) ProveSpace(challenges []int64) ([][]byte, [][][]byte, [][][]byte, [][][][]byte) {
	hashes := make([][]byte, len(challenges))
	proofs := make([][][]byte, len(challenges))
	parents := make([][][]byte, len(challenges))
	pProofs := make([][][][]byte, len(challenges))
	for i := range challenges {
		hashes[i], proofs[i] = p.Open(challenges[i])
		ps := p.graph.GetParents(challenges[i])
		for _, parent := range ps {
			if parent != -1 {
				hash, proof := p.Open(parent)
				parents[i] = append(parents[i], hash)
				pProofs[i] = append(pProofs[i], proof)
			}
		}
	}
	return hashes, parents, proofs, pProofs
}
