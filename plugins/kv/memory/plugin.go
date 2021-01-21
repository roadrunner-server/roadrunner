package memory

import (
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

// PluginName is user friendly name for the plugin
const PluginName = "memory"

type Plugin struct {
	// heap is user map for the key-value pairs
	heap sync.Map
	stop chan struct{}

	log logger.Logger
	cfg *Config
}

func (s *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("in_memory_plugin_init")
	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}
	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, err)
	}

	s.cfg.InitDefaults()
	s.log = log

	s.stop = make(chan struct{}, 1)
	return nil
}

func (s *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	// start in-memory gc for kv
	go s.gc()

	return errCh
}

func (s *Plugin) Stop() error {
	const op = errors.Op("in_memory_plugin_stop")
	err := s.Close()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *Plugin) Has(keys ...string) (map[string]bool, error) {
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

func (s *Plugin) Get(key string) ([]byte, error) {
	const op = errors.Op("in_memory_plugin_get")
	// to get cases like "  "
	keyTrimmed := strings.TrimSpace(key)
	if keyTrimmed == "" {
		return nil, errors.E(op, errors.EmptyKey)
	}

	if data, exist := s.heap.Load(key); exist {
		// here might be a panic
		// but data only could be a string, see Set function
		return []byte(data.(kv.Item).Value), nil
	}
	return nil, nil
}

func (s *Plugin) MGet(keys ...string) (map[string]interface{}, error) {
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

	m := make(map[string]interface{}, len(keys))

	for i := range keys {
		if value, ok := s.heap.Load(keys[i]); ok {
			m[keys[i]] = value.(kv.Item).Value
		}
	}

	return m, nil
}

func (s *Plugin) Set(items ...kv.Item) error {
	const op = errors.Op("in_memory_plugin_set")
	if items == nil {
		return errors.E(op, errors.NoKeys)
	}

	for i := range items {
		// TTL is set
		if items[i].TTL != "" {
			// check the TTL in the item
			_, err := time.Parse(time.RFC3339, items[i].TTL)
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
func (s *Plugin) MExpire(items ...kv.Item) error {
	const op = errors.Op("in_memory_plugin_mexpire")
	for i := range items {
		if items[i].TTL == "" || strings.TrimSpace(items[i].Key) == "" {
			return errors.E(op, errors.Str("should set timeout and at least one key"))
		}

		// if key exist, overwrite it value
		if pItem, ok := s.heap.Load(items[i].Key); ok {
			// check that time is correct
			_, err := time.Parse(time.RFC3339, items[i].TTL)
			if err != nil {
				return errors.E(op, err)
			}
			tmp := pItem.(kv.Item)
			// guess that t is in the future
			// in memory is just FOR TESTING PURPOSES
			// LOGIC ISN'T IDEAL
			s.heap.Store(items[i].Key, kv.Item{
				Key:   items[i].Key,
				Value: tmp.Value,
				TTL:   items[i].TTL,
			})
		}
	}

	return nil
}

func (s *Plugin) TTL(keys ...string) (map[string]interface{}, error) {
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

	m := make(map[string]interface{}, len(keys))

	for i := range keys {
		if item, ok := s.heap.Load(keys[i]); ok {
			m[keys[i]] = item.(kv.Item).TTL
		}
	}
	return m, nil
}

func (s *Plugin) Delete(keys ...string) error {
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

// Close clears the in-memory storage
func (s *Plugin) Close() error {
	s.stop <- struct{}{}
	return nil
}

// RPCService returns associated rpc service.
func (s *Plugin) RPC() interface{} {
	return kv.NewRPCServer(s, s.log)
}

// Name returns plugin user-friendly name
func (s *Plugin) Name() string {
	return PluginName
}

// ================================== PRIVATE ======================================

func (s *Plugin) gc() {
	// TODO check
	ticker := time.NewTicker(time.Duration(s.cfg.Interval) * time.Second)
	for {
		select {
		case <-s.stop:
			ticker.Stop()
			return
		case now := <-ticker.C:
			// check every second
			s.heap.Range(func(key, value interface{}) bool {
				v := value.(kv.Item)
				if v.TTL == "" {
					return true
				}

				t, err := time.Parse(time.RFC3339, v.TTL)
				if err != nil {
					return false
				}

				if now.After(t) {
					s.log.Debug("key deleted", "key", key)
					s.heap.Delete(key)
				}
				return true
			})
		}
	}
}
