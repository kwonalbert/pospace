package pospace

import (
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"testing"
	"time"
)

//exp* gets setup in test.go
var prover *Prover = nil
var verifier *Verifier = nil
var pk []byte
var index int64 = 3
var beta int = 1
var graphDir string = "posgraph/test"

func TestPoS(t *testing.T) {
	seed := make([]byte, 64)
	rand.Read(seed)
	challenges := verifier.SelectChallenges(seed)
	now := time.Now()
	hashes, parents, proofs, pProofs := prover.ProveSpace(challenges)
	fmt.Printf("Prove: %f\n", time.Since(now).Seconds())

	now = time.Now()
	if !verifier.VerifySpace(challenges, hashes, parents, proofs, pProofs) {
		log.Fatal("Verify space failed:", challenges)
	}
	fmt.Printf("Verify: %f\n", time.Since(now).Seconds())
}

func TestMain(m *testing.M) {
	pk = []byte{1}

	runtime.GOMAXPROCS(runtime.NumCPU())

	id := flag.Int("index", 1, "graph index")
	flag.Parse()
	index = int64(*id)

	prover = NewProver(pk, index, graphDir, ".")

	now := time.Now()
	commit := prover.Init()
	fmt.Printf("%d. Graph commit: %fs\n", index, time.Since(now).Seconds())

	root := commit.Commit
	verifier = NewVerifier(pk, index, beta, root, graphDir)

	os.Exit(m.Run())
}
