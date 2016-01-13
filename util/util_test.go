package util

import (
	"log"
	"math/big"
	"testing"
)

func TestPow(t *testing.T) {
	x := big.NewFloat(0.12381245613960218386)
	n := int64(3)
	res := Pow(x, n)
	exp := big.NewFloat(0.00189798605)
	diff := new(big.Float).Sub(res, exp)
	diff = diff.Abs(diff)
	if diff.Cmp(big.NewFloat(0.00000001)) >= 0 {
		log.Fatal("Pow failed:", exp, res)
	}
}

func TestRoot(t *testing.T) {
	x := big.NewFloat(0.12381245613960218386)
	n := int64(16)
	res := Root(x, n)
	exp := big.NewFloat(0.8776023372475015)
	diff := new(big.Float).Sub(res, exp)
	diff = diff.Abs(diff)
	if diff.Cmp(big.NewFloat(0.00000001)) >= 0 {
		log.Fatal("Exp failed:", exp, res)
	}
}

func TestUnion(t *testing.T) {
	l1 := []int64{1, 2, 3, 5, 6}
	l2 := []int64{2, 3, 4, 5}
	exp := []int64{1, 2, 3, 4, 5, 6}
	res := Union(l1, l2)
	log.Println(exp, res)
}
