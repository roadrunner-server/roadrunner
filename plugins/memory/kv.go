package memory

import (
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	kvv1 "github.com/spiral/roadrunner/v2/proto/kv/v1beta"
)

type Driver struct {
	clearMu sync.RWMutex
	heap    sync.Map
	// stop is used to stop keys GC and close boltdb connection
	stop chan struct{}
	log  logger.Logger
	cfg  *Config
}

func NewInMemoryDriver(log logger.Logger, key string, cfgPlugin config.Configurer, stop chan struct{}) (kv.Storage, error) {
	const op = errors.Op("new_in_memory_driver")

	d := &Driver{
		stop: stop,
		log:  log,
	}

	err := cfgPlugin.UnmarshalKey(key, &d.cfg)
	if err != nil {
		return nil, errors.E(op, err)
	}

	d.cfg.InitDefaults()

	go d.gc()

	return d, nil
}

func (s *Driver) Has(keys ...string) (map[string]bool, error) {
	const op = errors.Op("in_memory_plugin_has")
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}
	m := make(map[string]bool)
	for i := range keys {
		keyTrimmed := strings.TrimSpace(keys[i])
		if keyTrimmed == "" {
			return nil, errors.E(op, errors.EmptyKey)
		}

		if _, ok := s.heap.Load(keys[i]); ok {
			m[keys[i]] = true
		}
	}

	return m, nil
}

func (s *Driver) Get(key string) ([]byte, error) {
	const op = errors.Op("in_memory_plugin_get")
	// to get cases like "  "
	keyTrimmed := strings.TrimSpace(key)
	if keyTrimmed == "" {
		return nil, errors.E(op, errors.EmptyKey)
	}

	if data, exist := s.heap.Load(key); exist {
		// here might be a panic
		// but data only could be a string, see Set function
		return data.(*kvv1.Item).Value, nil
	}
	return nil, nil
}

func (s *Driver) MGet(keys ...string) (map[string][]byte, error) {
	const op = errors.Op("in_memory_plugin_mget")
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}

	// should not be empty keys
	for i := range keys {
		keyTrimmed := strings.TrimSpace(keys[i])
		if keyTrimmed == "" {
			return nil, errors.E(op, errors.EmptyKey)
		}
	}

	m := make(map[string][]byte, len(keys))

	for i := range keys {
		if value, ok := s.heap.Load(keys[i]); ok {
			m[keys[i]] = value.(*kvv1.Item).Value
		}
	}

	return m, nil
}

func (s *Driver) Set(items ...*kvv1.Item) error {
	const op = errors.Op("in_memory_plugin_set")
	if items == nil {
		return errors.E(op, errors.NoKeys)
	}

	for i := range items {
		if items[i] == nil {
			continue
		}
		// TTL is set
		if items[i].Timeout != "" {
			// check the TTL in the item
			_, err := time.Parse(time.RFC3339, items[i].Timeout)
			if err != nil {
				return err
			}
		}

		s.heap.Store(items[i].Key, items[i])
	}
	return nil
}

// MExpire sets the expiration time to the key
// If key already has the expiration time, it will be overwritten
func (s *Driver) MExpire(items ...*kvv1.Item) error {
	const op = errors.Op("in_memory_plugin_mexpire")
	for i := range items {
		if items[i] == nil {
			continue
		}
		if items[i].Timeout == "" || strings.TrimSpace(items[i].Key) == "" {
			return errors.E(op, errors.Str("should set timeout and at least one key"))
		}

		// if key exist, overwrite it value
		if pItem, ok := s.heap.LoadAndDelete(items[i].Key); ok {
			// check that time is correct
			_, err := time.Parse(time.RFC3339, items[i].Timeout)
			if err != nil {
				return errors.E(op, err)
			}
			tmp := pItem.(*kvv1.Item)
			// guess that t is in the future
			// in memory is just FOR TESTING PURPOSES
			// LOGIC ISN'T IDEAL
			s.heap.Store(items[i].Key, &kvv1.Item{
				Key:     items[i].Key,
				Value:   tmp.Value,
				Timeout: items[i].Timeout,
			})
		}
	}

	return nil
}

func (s *Driver) TTL(keys ...string) (map[string]string, error) {
	const op = errors.Op("in_memory_plugin_ttl")
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}

	// should not be empty keys
	for i := range keys {
		keyTrimmed := strings.TrimSpace(keys[i])
		if keyTrimmed == "" {
			return nil, errors.E(op, errors.EmptyKey)
		}
	}

	m := make(map[string]string, len(keys))

	for i := range keys {
		if item, ok := s.heap.Load(keys[i]); ok {
			m[keys[i]] = item.(*kvv1.Item).Timeout
		}
	}
	return m, nil
}

func (s *Driver) Delete(keys ...string) error {
	const op = errors.Op("in_memory_plugin_delete")
	if keys == nil {
		return errors.E(op, errors.NoKeys)
	}

	// should not be empty keys
	for i := range keys {
		keyTrimmed := strings.TrimSpace(keys[i])
		if keyTrimmed == "" {
			return errors.E(op, errors.EmptyKey)
		}
	}

	for i := range keys {
		s.heap.Delete(keys[i])
	}
	return nil
}

func (s *Driver) Clear() error {
	s.clearMu.Lock()
	s.heap = sync.Map{}
	s.clearMu.Unlock()

	return nil
}

// ================================== PRIVATE ======================================

func (s *Driver) gc() {
	ticker := time.NewTicker(time.Duration(s.cfg.Interval) * time.Second)
	for {
		select {
		case <-s.stop:
			ticker.Stop()
			return
		case now := <-ticker.C:
			// mutes needed to clear the map
			s.clearMu.RLock()

			// check every second
			s.heap.Range(func(key, value interface{}) bool {
				v := value.(*kvv1.Item)
				if v.Timeout == "" {
					return true
				}

				t, err := time.Parse(time.RFC3339, v.Timeout)
				if err != nil {
					return false
				}

				if now.After(t) {
					s.log.Debug("key deleted", "key", key)
					s.heap.Delete(key)
				}
				return true
			})

			s.clearMu.RUnlock()
		}
	}
}
