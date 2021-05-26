package bst

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewBST(t *testing.T) {
	// create a new bst
	g := NewBST()

	for i := 0; i < 100; i++ {
		g.Insert(uuid.NewString(), "comments")
	}

	for i := 0; i < 100; i++ {
		g.Insert(uuid.NewString(), "comments2")
	}

	for i := 0; i < 100; i++ {
		g.Insert(uuid.NewString(), "comments3")
	}

	// should be 100
	exist := g.Get("comments")
	assert.Len(t, exist, 100)

	// should be 100
	exist2 := g.Get("comments2")
	assert.Len(t, exist2, 100)

	// should be 100
	exist3 := g.Get("comments3")
	assert.Len(t, exist3, 100)
}
