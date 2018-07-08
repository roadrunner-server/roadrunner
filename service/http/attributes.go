package http

import (
	"context"
	"net/http"
)

const contextKey = "psr:attributes"

type attrs map[string]interface{}

// InitAttributes returns request with new context and attribute bag.
func InitAttributes(r *http.Request) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKey, attrs{}))
}

// AllAttributes returns all context attributes.
func AllAttributes(r *http.Request) map[string]interface{} {
	v := r.Context().Value(contextKey)
	if v == nil {
		return nil
	}

	return v.(attrs)
}

// Get gets the value from request context. It replaces any existing
// values.
func GetAttribute(r *http.Request, key string) interface{} {
	v := r.Context().Value(contextKey)
	if v == nil {
		return ""
	}

	return v.(attrs).Get(key)
}

// Set sets the key to value. It replaces any existing
// values. Context specific.
func SetAttribute(r *http.Request, key string, value interface{}) {
	v := r.Context().Value(contextKey)
	v.(attrs).Set(key, value)
}

// Get gets the value associated with the given key.
func (v attrs) Get(key string) interface{} {
	if v == nil {
		return ""
	}

	return v[key]
}

// Set sets the key to value. It replaces any existing
// values.
func (v attrs) Set(key string, value interface{}) {
	v[key] = value
}

// Del deletes the value associated with key.
func (v attrs) Del(key string) {
	delete(v, key)
}