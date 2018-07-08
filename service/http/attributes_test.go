package http

import (
	"testing"
	"net/http"
	"github.com/stretchr/testify/assert"
)

func TestAllAttributes(t *testing.T) {
	r := &http.Request{}
	r = InitAttributes(r)

	SetAttribute(r, "key", "value")

	assert.Equal(t, AllAttributes(r), map[string]interface{}{
		"key": "value",
	})
}

func TestAllAttributesNone(t *testing.T) {
	r := &http.Request{}
	r = InitAttributes(r)

	assert.Equal(t, AllAttributes(r), map[string]interface{}{})
}

func TestAllAttributesNone2(t *testing.T) {
	r := &http.Request{}

	assert.Equal(t, AllAttributes(r), map[string]interface{}{})
}

func TestGetAttribute(t *testing.T) {
	r := &http.Request{}
	r = InitAttributes(r)

	SetAttribute(r, "key", "value")
	assert.Equal(t, GetAttribute(r, "key"), "value")
}

func TestGetAttributeNone(t *testing.T) {
	r := &http.Request{}
	r = InitAttributes(r)

	assert.Equal(t, GetAttribute(r, "key"), nil)
}

func TestGetAttributeNone2(t *testing.T) {
	r := &http.Request{}

	assert.Equal(t, GetAttribute(r, "key"), nil)
}

func TestSetAttribute(t *testing.T) {
	r := &http.Request{}
	r = InitAttributes(r)

	SetAttribute(r, "key", "value")
	assert.Equal(t, GetAttribute(r, "key"), "value")
}

func TestSetAttributeNone(t *testing.T) {
	r := &http.Request{}

	SetAttribute(r, "key", "value")
	assert.Equal(t, GetAttribute(r, "key"), nil)
}