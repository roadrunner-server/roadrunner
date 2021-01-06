package kv

// Item represents general storage item
type Item struct {
	// Key of item
	Key string
	// Value of item
	Value string
	// live until time provided by TTL in RFC 3339 format
	TTL string
}

// Storage represents single abstract storage.
type Storage interface {
	// Has checks if value exists.
	Has(keys ...string) (map[string]bool, error)

	// Get loads value content into a byte slice.
	Get(key string) ([]byte, error)

	// MGet loads content of multiple values
	// Returns the map with existing keys and associated values
	MGet(keys ...string) (map[string]interface{}, error)

	// Set used to upload item to KV with TTL
	// 0 value in TTL means no TTL
	Set(items ...Item) error

	// MExpire sets the TTL for multiply keys
	MExpire(items ...Item) error

	// TTL return the rest time to live for provided keys
	// Not supported for the memcached and boltdb
	TTL(keys ...string) (map[string]interface{}, error)

	// Delete one or multiple keys.
	Delete(keys ...string) error

	// Close closes the storage and underlying resources.
	Close() error
}
