package attributes

import (
	"context"
	"errors"
	"net/http"
)

// contextKey is a value for use with context.WithValue. It's used as
// a pointer so it fits in an interface{} without allocation.
type contextKey struct {
	name string
}

func (k *contextKey) String() string { return k.name }

var (
	// PsrContextKey is a context key. It can be used in the http attributes
	PsrContextKey = &contextKey{"psr_attributes"}
)

type attrs map[string]interface{}

func (v attrs) get(key string) interface{} {
	if v == nil {
		return ""
	}

	return v[key]
}

func (v attrs) set(key string, value interface{}) {
	v[key] = value
}

func (v attrs) del(key string) {
	delete(v, key)
}

// Init returns request with new context and attribute bag.
func Init(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), PsrContextKey, attrs{}))
}

// All returns all context attributes.
func All(r *http.Request) map[string]interface{} {
	v := r.Context().Value(PsrContextKey)
	if v == nil {
		return attrs{}
	}

	return v.(attrs)
}

// Get gets the value from request context. It replaces any existing
// values.
func Get(r *http.Request, key string) interface{} {
	v := r.Context().Value(PsrContextKey)
	if v == nil {
		return nil
	}

	return v.(attrs).get(key)
}

// Set sets the key to value. It replaces any existing
// values. Context specific.
func Set(r *http.Request, key string, value interface{}) error {
	v := r.Context().Value(PsrContextKey)
	if v == nil {
		return errors.New("unable to find `psr:attributes` context key")
	}

	v.(attrs).set(key, value)
	return nil
}

// Delete deletes values associated with attribute key.
func (v attrs) Delete(key string) {
	if v == nil {
		return
	}

	v.del(key)
}
