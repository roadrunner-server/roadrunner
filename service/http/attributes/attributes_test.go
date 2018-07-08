package attributes

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAllAttributes(t *testing.T) {
	r := &http.Request{}
	r = Init(r)

	Set(r, "key", "value")

	assert.Equal(t, All(r), map[string]interface{}{
		"key": "value",
	})
}

func TestAllAttributesNone(t *testing.T) {
	r := &http.Request{}
	r = Init(r)

	assert.Equal(t, All(r), map[string]interface{}{})
}

func TestAllAttributesNone2(t *testing.T) {
	r := &http.Request{}

	assert.Equal(t, All(r), map[string]interface{}{})
}

func TestGetAttribute(t *testing.T) {
	r := &http.Request{}
	r = Init(r)

	Set(r, "key", "value")
	assert.Equal(t, Get(r, "key"), "value")
}

func TestGetAttributeNone(t *testing.T) {
	r := &http.Request{}
	r = Init(r)

	assert.Equal(t, Get(r, "key"), nil)
}

func TestGetAttributeNone2(t *testing.T) {
	r := &http.Request{}

	assert.Equal(t, Get(r, "key"), nil)
}

func TestSetAttribute(t *testing.T) {
	r := &http.Request{}
	r = Init(r)

	Set(r, "key", "value")
	assert.Equal(t, Get(r, "key"), "value")
}

func TestSetAttributeNone(t *testing.T) {
	r := &http.Request{}

	Set(r, "key", "value")
	assert.Equal(t, Get(r, "key"), nil)
}
