package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
)

const PluginName = "memory"

type Plugin struct {
	heap *sync.Map
	stop chan struct{}

	log logger.Logger
	cfg *Config
}

func NewInMemoryStorage() kv.Storage {
	p := &Plugin{
		heap: &sync.Map{},
		stop: make(chan struct{}),
	}

	go p.gc()

	return p
}

func (s *Plugin) Init(cfg config.Configurer, log logger.Logger) error {
	const op = errors.Op("in-memory storage init")
	s.cfg = &Config{}
	s.cfg.InitDefaults()

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, err)
	}
	s.log = log
	// init in-memory
	s.heap = &sync.Map{}
	s.stop = make(chan struct{}, 1)
	return nil
}

func (s Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	// start in-memory gc for kv
	go s.gc()

	return errCh
}

func (s Plugin) Stop() error {
	const op = errors.Op("in-memory storage stop")
	err := s.Close()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s Plugin) Has(ctx context.Context, keys ...string) (map[string]bool, error) {
	const op = errors.Op("in-memory storage Has")
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}
	m := make(map[string]bool)
	for _, key := range keys {
		keyTrimmed := strings.TrimSpace(key)
		if keyTrimmed == "" {
			return nil, errors.E(op, errors.EmptyKey)
		}

		if _, ok := s.heap.Load(key); ok {
			m[key] = true
		}
	}

	return m, nil
}

func (s Plugin) Get(ctx context.Context, key string) ([]byte, error) {
	const op = errors.Op("in-memory storage Get")
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

func (s Plugin) MGet(ctx context.Context, keys ...string) (map[string]interface{}, error) {
	const op = errors.Op("in-memory storage MGet")
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
		if value, ok := s.heap.Load(key); ok {
			m[key] = value
		}
	}

	return m, nil
}

func (s Plugin) Set(ctx context.Context, items ...kv.Item) error {
	const op = errors.Op("in-memory storage Set")
	if items == nil {
		return errors.E(op, errors.NoKeys)
	}

	for _, item := range items {
		// TTL is set
		if item.TTL != "" {
			// check the TTL in the item
			_, err := time.Parse(time.RFC3339, item.TTL)
			if err != nil {
				return err
			}
		}

		s.heap.Store(item.Key, item)
	}
	return nil
}

// MExpire sets the expiration time to the key
// If key already has the expiration time, it will be overwritten
func (s Plugin) MExpire(ctx context.Context, items ...kv.Item) error {
	const op = errors.Op("in-memory storage MExpire")
	for _, item := range items {
		if item.TTL == "" || strings.TrimSpace(item.Key) == "" {
			return errors.E(op, errors.Str("should set timeout and at least one key"))
		}

		// if key exist, overwrite it value
		if pItem, ok := s.heap.Load(item.Key); ok {
			// check that time is correct
			_, err := time.Parse(time.RFC3339, item.TTL)
			if err != nil {
				return errors.E(op, err)
			}
			tmp := pItem.(kv.Item)
			// guess that t is in the future
			// in memory is just FOR TESTING PURPOSES
			// LOGIC ISN'T IDEAL
			s.heap.Store(item.Key, kv.Item{
				Key:   item.Key,
				Value: tmp.Value,
				TTL:   item.TTL,
			})
		}
	}

	return nil
}

func (s Plugin) TTL(ctx context.Context, keys ...string) (map[string]interface{}, error) {
	const op = errors.Op("in-memory storage TTL")
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
		if item, ok := s.heap.Load(key); ok {
			m[key] = item.(kv.Item).TTL
		}
	}
	return m, nil
}

func (s Plugin) Delete(ctx context.Context, keys ...string) error {
	const op = errors.Op("in-memory storage Delete")
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
		s.heap.Delete(key)
	}
	return nil
}

// Close clears the in-memory storage
func (s Plugin) Close() error {
	s.heap = &sync.Map{}
	s.stop <- struct{}{}
	return nil
}

// ================================== PRIVATE ======================================

func (s *Plugin) gc() {
	// TODO check
	ticker := time.NewTicker(time.Millisecond * 500)
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
					s.heap.Delete(key)
				}
				return true
			})
		}
	}
}
