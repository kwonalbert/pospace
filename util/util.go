package util

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
	"sort"
)

// return: x^y
func Pow(x *big.Float, n int64) *big.Float {
	res := new(big.Float).Copy(x)
	if n < 0 {
		res = res.Quo(big.NewFloat(1), res)
		n = -n
	} else if n == 0 {
		return big.NewFloat(1)
	}
	y := big.NewFloat(1)
	for i := n; i > 1; {
		if i%2 == 0 {
			i /= 2
		} else {
			y = y.Mul(res, y)
			i = (i - 1) / 2
		}
		res = res.Mul(res, res)
	}
	return res.Mul(res, y)
}

// Implements the nth root algorithm from
// https://en.wikipedia.org/wiki/Nth_root_algorithm
// return: nth root of x within some epsilon
func Root(x *big.Float, n int64) *big.Float {
	guess := new(big.Float).Quo(x, big.NewFloat(float64(n)))
	diff := big.NewFloat(1)
	ep := big.NewFloat(0.00000001)
	abs := new(big.Float).Abs(diff)
	for abs.Cmp(ep) >= 0 {
		//fmt.Println(guess, abs)
		prev := Pow(guess, n-1)
		diff = new(big.Float).Quo(x, prev)
		diff = diff.Sub(diff, guess)
		diff = diff.Quo(diff, big.NewFloat(float64(n)))

		guess = guess.Add(guess, diff)
		abs = new(big.Float).Abs(diff)
	}
	return guess
}

//return: floor log base 2 of x
func Log2(x int64) int64 {
	var r int64 = 0
	for ; x > 1; x >>= 1 {
		r++
	}
	return r
}

//From hackers delight
func Count(x uint64) int {
	x = x - ((x >> 1) & 0x55555555)
	x = (x & 0x33333333) + ((x >> 2) & 0x33333333)
	x = (x + (x >> 4)) & 0x0F0F0F0F
	x = x + (x >> 8)
	x = x + (x >> 16)
	return int(x & 0x0000003F)
}

func Subtree(log2, node int64) int64 {
	level := (log2 + 1) - Log2(node)
	return int64((1 << uint64(level)) - 1)
}

// post-order is better for disk than bfs
func BfsToPost(pow2, log2, node int64) int64 {
	if node == 0 {
		return 0
	}
	cur := node
	res := int64(0)
	for cur != 1 {
		if cur%2 == 0 {
			res -= (Subtree(log2, cur) + 1)
		} else {
			res--
		}
		cur /= 2
	}
	res += 2*pow2 - 1
	return res
}

func Min(x, y int64) int64 {
	if x < y {
		return x
	}
	return y
}

func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

type int64arr []int64

func (a int64arr) Len() int           { return len(a) }
func (a int64arr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a int64arr) Less(i, j int) bool { return a[i] < a[j] }

// return n values that ranges [l, u-1], sorted
func NRandRange(l, u, n int64) []int64 {
	seen := make([]bool, u-l)
	var vals int64arr = make([]int64, n)
	for i := range seen {
		seen[i] = false
	}
	count := int64(0)
	for count < n {
		buf := make([]byte, 8)
		_, err := rand.Read(buf)
		if err != nil {
			panic(err)
		}
		v, _ := binary.Uvarint(buf)
		val := int64(v % (uint64(u - l)))
		if !seen[val] {
			seen[val] = true
			vals[count] = val + l
			count++
		}
	}
	sort.Sort(vals)
	return vals
}

func Union(l1, l2 []int64) []int64 {
	seen := make(map[int64]bool)
	var u []int64
	i := 0
	j := 0
	for i < len(l1) && j < len(l2) {
		var add int64
		if l1[i] <= l2[j] {
			add = l1[i]
			i++
		} else {
			add = l2[j]
			j++
		}
		if _, ok := seen[add]; !ok {
			u = append(u, add)
			seen[add] = true
		}
	}
	for i < len(l1) {
		if _, ok := seen[l1[i]]; !ok {
			u = append(u, l1[i])
			seen[l1[i]] = true
		}
		i++
	}
	for j < len(l2) {
		if _, ok := seen[l2[j]]; !ok {
			u = append(u, l2[j])
			seen[l2[j]] = true
		}
		j++
	}
	return u
}
