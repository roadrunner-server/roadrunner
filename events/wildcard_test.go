package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWildcard(t *testing.T) {
	w, err := newWildcard("http.*")
	assert.NoError(t, err)
	assert.True(t, w.match("http.SuperEvent"))
	assert.False(t, w.match("https.SuperEvent"))
	assert.False(t, w.match(""))
	assert.False(t, w.match("*"))
	assert.False(t, w.match("****"))
	assert.True(t, w.match("http.****"))

	// *.* -> *
	w, err = newWildcard("*")
	assert.NoError(t, err)
	assert.True(t, w.match("http.SuperEvent"))
	assert.True(t, w.match("https.SuperEvent"))
	assert.True(t, w.match(""))
	assert.True(t, w.match("*"))
	assert.True(t, w.match("****"))
	assert.True(t, w.match("http.****"))

	w, err = newWildcard("*.WorkerError")
	assert.NoError(t, err)
	assert.False(t, w.match("http.SuperEvent"))
	assert.False(t, w.match("https.SuperEvent"))
	assert.False(t, w.match(""))
	assert.False(t, w.match("*"))
	assert.False(t, w.match("****"))
	assert.False(t, w.match("http.****"))
	assert.True(t, w.match("http.WorkerError"))

	w, err = newWildcard("http.WorkerError")
	assert.NoError(t, err)
	assert.False(t, w.match("http.SuperEvent"))
	assert.False(t, w.match("https.SuperEvent"))
	assert.False(t, w.match(""))
	assert.False(t, w.match("*"))
	assert.False(t, w.match("****"))
	assert.False(t, w.match("http.****"))
	assert.True(t, w.match("http.WorkerError"))

	w, err = newWildcard("http.Worker*")
	assert.NoError(t, err)
	assert.True(t, w.match("http.WorkerFoo"))
	assert.False(t, w.match("h*.SuperEvent"))
	assert.False(t, w.match("h*.Worker"))
}
