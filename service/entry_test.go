package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntry_CanServeFalse(t *testing.T) {
	e := &entry{svc: nil}
	assert.False(t, e.canServe())
}

func TestEntry_CanServeTrue(t *testing.T) {
	e := &entry{svc: &testService{}}
	assert.True(t, e.canServe())
}
