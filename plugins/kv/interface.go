package kv

import kvv1 "github.com/spiral/roadrunner/v2/pkg/proto/kv/v1beta"

// Storage represents single abstract storage.
type Storage interface {
	// Has checks if value exists.
	Has(keys ...string) (map[string]bool, error)

	// Get loads value content into a byte slice.
	Get(key string) ([]byte, error)

	// MGet loads content of multiple values
	// Returns the map with existing keys and associated values
	MGet(keys ...string) (map[string][]byte, error)

	// Set used to upload item to KV with TTL
	// 0 value in TTL means no TTL
	Set(items ...*kvv1.Item) error

	// MExpire sets the TTL for multiply keys
	MExpire(items ...*kvv1.Item) error

	// TTL return the rest time to live for provided keys
	// Not supported for the memcached and boltdb
	TTL(keys ...string) (map[string]string, error)

	// Delete one or multiple keys.
	Delete(keys ...string) error
}

// StorageDriver interface provide storage
type StorageDriver interface {
	Provider
}

// Provider provides storage based on the config
type Provider interface {
	// Provide provides Storage based on the config key
	Provide(key string) (Storage, error)
}
