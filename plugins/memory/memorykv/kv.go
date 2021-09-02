package memorykv

import (
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
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

func NewInMemoryDriver(key string, log logger.Logger, cfgPlugin config.Configurer) (*Driver, error) {
	const op = errors.Op("new_in_memory_driver")

	d := &Driver{
		stop: make(chan struct{}),
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

func (d *Driver) Has(keys ...string) (map[string]bool, error) {
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

		if _, ok := d.heap.Load(keys[i]); ok {
			m[keys[i]] = true
		}
	}

	return m, nil
}

func (d *Driver) Get(key string) ([]byte, error) {
	const op = errors.Op("in_memory_plugin_get")
	// to get cases like "  "
	keyTrimmed := strings.TrimSpace(key)
	if keyTrimmed == "" {
		return nil, errors.E(op, errors.EmptyKey)
	}

	if data, exist := d.heap.Load(key); exist {
		// here might be a panic
		// but data only could be a string, see Set function
		return data.(*kvv1.Item).Value, nil
	}
	return nil, nil
}

func (d *Driver) MGet(keys ...string) (map[string][]byte, error) {
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
		if value, ok := d.heap.Load(keys[i]); ok {
			m[keys[i]] = value.(*kvv1.Item).Value
		}
	}

	return m, nil
}

func (d *Driver) Set(items ...*kvv1.Item) error {
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

		d.heap.Store(items[i].Key, items[i])
	}
	return nil
}

// MExpire sets the expiration time to the key
// If key already has the expiration time, it will be overwritten
func (d *Driver) MExpire(items ...*kvv1.Item) error {
	const op = errors.Op("in_memory_plugin_mexpire")
	for i := range items {
		if items[i] == nil {
			continue
		}
		if items[i].Timeout == "" || strings.TrimSpace(items[i].Key) == "" {
			return errors.E(op, errors.Str("should set timeout and at least one key"))
		}

		// if key exist, overwrite it value
		if pItem, ok := d.heap.LoadAndDelete(items[i].Key); ok {
			// check that time is correct
			_, err := time.Parse(time.RFC3339, items[i].Timeout)
			if err != nil {
				return errors.E(op, err)
			}
			tmp := pItem.(*kvv1.Item)
			// guess that t is in the future
			// in memory is just FOR TESTING PURPOSES
			// LOGIC ISN'T IDEAL
			d.heap.Store(items[i].Key, &kvv1.Item{
				Key:     items[i].Key,
				Value:   tmp.Value,
				Timeout: items[i].Timeout,
			})
		}
	}

	return nil
}

func (d *Driver) TTL(keys ...string) (map[string]string, error) {
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
		if item, ok := d.heap.Load(keys[i]); ok {
			m[keys[i]] = item.(*kvv1.Item).Timeout
		}
	}
	return m, nil
}

func (d *Driver) Delete(keys ...string) error {
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
		d.heap.Delete(keys[i])
	}
	return nil
}

func (d *Driver) Clear() error {
	d.clearMu.Lock()
	d.heap = sync.Map{}
	d.clearMu.Unlock()

	return nil
}

func (d *Driver) Stop() {
	d.stop <- struct{}{}
}

// ================================== PRIVATE ======================================

func (d *Driver) gc() {
	ticker := time.NewTicker(time.Duration(d.cfg.Interval) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.stop:
			return
		case now := <-ticker.C:
			// mutes needed to clear the map
			d.clearMu.RLock()

			// check every second
			d.heap.Range(func(key, value interface{}) bool {
				v := value.(*kvv1.Item)
				if v.Timeout == "" {
					return true
				}

				t, err := time.Parse(time.RFC3339, v.Timeout)
				if err != nil {
					return false
				}

				if now.After(t) {
					d.log.Debug("key deleted", "key", key)
					d.heap.Delete(key)
				}
				return true
			})

			d.clearMu.RUnlock()
		}
	}
}
