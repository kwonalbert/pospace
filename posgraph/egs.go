package posgraph

import (
	// "encoding/binary"
	// "golang.org/x/crypto/sha3"
	"os"
)

type EGSGraph struct {
	pk    []byte
	fn    string
	db    *os.File
	index int64
	log2  int64
	pow2  int64
	size  int64
}

type EGSNode struct {
	H []byte
}
