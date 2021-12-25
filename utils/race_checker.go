//go:build race

package utils

import (
	"crypto/sha512"
	"fmt"
	"runtime"
)

func SetChecker(b []byte) {
	if len(b) == 0 {
		return
	}
	c := checkIfConst(b)
	go c.isStillConst()
	runtime.SetFinalizer(c, (*constSlice).isStillConst)
}

type constSlice struct {
	b        []byte
	checksum [64]byte
}

func checkIfConst(b []byte) *constSlice {
	c := &constSlice{b: b}
	c.checksum = sha512.Sum512(c.b)
	return c
}

func (c *constSlice) isStillConst() {
	if sha512.Sum512(c.b) != c.checksum {
		panic(fmt.Sprintf("mutable access detected 0x%012x", &c.b[0]))
	}
}
