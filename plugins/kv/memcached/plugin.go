package memcached

import (
	"context"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "memcached"

var EmptyItem = kv.Item{}

type Plugin struct {
	// config
	cfg *Config
	// logger
	log logger.Logger
	// memcached client
	client *memcache.Client
}

// NewMemcachedClient returns a memcache client using the provided server(s)
// with equal weight. If a server is listed multiple times,
// it gets a proportional amount of weight.
func NewMemcachedClient(url string) kv.Storage {
	m := memcache.New(url)
	return &Plugin{
		client: m,
	}
}

func (s *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	const op = errors.Op("memcached init")
	s.cfg = &Config{}
	s.cfg.InitDefaults()
	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, err)
	}
	s.log = log
	return nil
}

func (s *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	s.client = memcache.New(s.cfg.Addr...)
	return errCh
}

// Memcached has no stop/close or smt similar to close the connection
func (s *Plugin) Stop() error {
	return nil
}

// Has checks the key for existence
func (s Plugin) Has(ctx context.Context, keys ...string) (map[string]bool, error) {
	const op = errors.Op("memcached Has")
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}
	m := make(map[string]bool, len(keys))
	for _, key := range keys {
		keyTrimmed := strings.TrimSpace(key)
		if keyTrimmed == "" {
			return nil, errors.E(op, errors.EmptyKey)
		}
		exist, err := s.client.Get(key)
		// ErrCacheMiss means that a Get failed because the item wasn't present.
		if err != nil && err != memcache.ErrCacheMiss {
			return nil, err
		}
		if exist != nil {
			m[key] = true
		}
	}
	return m, nil
}

// Get gets the item for the given key. ErrCacheMiss is returned for a
// memcache cache miss. The key must be at most 250 bytes in length.
func (s Plugin) Get(ctx context.Context, key string) ([]byte, error) {
	const op = errors.Op("memcached Get")
	// to get cases like "  "
	keyTrimmed := strings.TrimSpace(key)
	if keyTrimmed == "" {
		return nil, errors.E(op, errors.EmptyKey)
	}
	data, err := s.client.Get(key)
	// ErrCacheMiss means that a Get failed because the item wasn't present.
	if err != nil && err != memcache.ErrCacheMiss {
		return nil, err
	}
	if data != nil {
		// return the value by the key
		return data.Value, nil
	}
	// data is nil by some reason and error also nil
	return nil, nil
}

// return map with key -- string
// and map value as value -- []byte
func (s Plugin) MGet(ctx context.Context, keys ...string) (map[string]interface{}, error) {
	const op = errors.Op("memcached MGet")
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}

	// should not be empty keys
	for _, key := range keys {
		keyTrimmed := strings.TrimSpace(key)
		if keyTrimmed == "" {
			return nil, errors.E(op, errors.EmptyKey)
		}
	}

	m := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		// Here also MultiGet
		data, err := s.client.Get(key)
		// ErrCacheMiss means that a Get failed because the item wasn't present.
		if err != nil && err != memcache.ErrCacheMiss {
			return nil, err
		}
		if data != nil {
			m[key] = data.Value
		}
	}

	return m, nil
}

// Set sets the KV pairs. Keys should be 250 bytes maximum
// TTL:
// Expiration is the cache expiration time, in seconds: either a relative
// time from now (up to 1 month), or an absolute Unix epoch time.
// Zero means the Item has no expiration time.
func (s Plugin) Set(ctx context.Context, items ...kv.Item) error {
	const op = errors.Op("memcached Set")
	if items == nil {
		return errors.E(op, errors.NoKeys)
	}

	for i := range items {
		if items[i] == EmptyItem {
			return errors.E(op, errors.EmptyItem)
		}

		// pre-allocate item
		memcachedItem := &memcache.Item{
			Key: items[i].Key,
			// unsafe convert
			Value: []byte(items[i].Value),
			Flags: 0,
		}

		// add additional TTL in case of TTL isn't empty
		if items[i].TTL != "" {
			// verify the TTL
			t, err := time.Parse(time.RFC3339, items[i].TTL)
			if err != nil {
				return err
			}
			memcachedItem.Expiration = int32(t.Unix())
		}

		err := s.client.Set(memcachedItem)
		if err != nil {
			return err
		}
	}

	return nil
}

// Expiration is the cache expiration time, in seconds: either a relative
// time from now (up to 1 month), or an absolute Unix epoch time.
// Zero means the Item has no expiration time.
func (s Plugin) MExpire(ctx context.Context, items ...kv.Item) error {
	const op = errors.Op("memcached MExpire")
	for _, item := range items {
		if item.TTL == "" || strings.TrimSpace(item.Key) == "" {
			return errors.E(op, errors.Str("should set timeout and at least one key"))
		}

		// verify provided TTL
		t, err := time.Parse(time.RFC3339, item.TTL)
		if err != nil {
			return err
		}

		// Touch updates the expiry for the given key. The seconds parameter is either
		// a Unix timestamp or, if seconds is less than 1 month, the number of seconds
		// into the future at which time the item will expire. Zero means the item has
		// no expiration time. ErrCacheMiss is returned if the key is not in the cache.
		// The key must be at most 250 bytes in length.
		err = s.client.Touch(item.Key, int32(t.Unix()))
		if err != nil {
			return err
		}
	}
	return nil
}

// return time in seconds (int32) for a given keys
func (s Plugin) TTL(ctx context.Context, keys ...string) (map[string]interface{}, error) {
	const op = errors.Op("memcached HTTLas")
	return nil, errors.E(op, errors.Str("not valid request for memcached, see https://github.com/memcached/memcached/issues/239"))
}

func (s Plugin) Delete(ctx context.Context, keys ...string) error {
	const op = errors.Op("memcached Has")
	if keys == nil {
		return errors.E(op, errors.NoKeys)
	}

	// should not be empty keys
	for _, key := range keys {
		keyTrimmed := strings.TrimSpace(key)
		if keyTrimmed == "" {
			return errors.E(op, errors.EmptyKey)
		}
	}

	for _, key := range keys {
		err := s.client.Delete(key)
		// ErrCacheMiss means that a Get failed because the item wasn't present.
		if err != nil && err != memcache.ErrCacheMiss {
			return err
		}
	}
	return nil
}

func (s Plugin) Close() error {
	return nil
}
