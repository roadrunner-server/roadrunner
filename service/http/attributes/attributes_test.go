package attributes

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllAttributes(t *testing.T) {
	r := &http.Request{}
	r = Init(r)

	err := Set(r, "key", "value")
	if err != nil {
		t.Errorf("error during the Set: error %v", err)
	}

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

	err := Set(r, "key", "value")
	if err != nil {
		t.Errorf("error during the Set: error %v", err)
	}
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

	err := Set(r, "key", "value")
	if err != nil {
		t.Errorf("error during the Set: error %v", err)
	}
	assert.Equal(t, Get(r, "key"), "value")
}

func TestSetAttributeNone(t *testing.T) {
	r := &http.Request{}

	err := Set(r, "key", "value")
	if err != nil {
		t.Errorf("error during the Set: error %v", err)
	}
	assert.Equal(t, Get(r, "key"), nil)
}
