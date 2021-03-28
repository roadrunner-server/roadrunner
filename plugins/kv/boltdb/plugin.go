package boltdb

import (
	"bytes"
	"encoding/gob"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	bolt "go.etcd.io/bbolt"
)

const PluginName = "boltdb"

// BoltDB K/V storage.
type Plugin struct {
	// db instance
	DB *bolt.DB
	// name should be UTF-8
	bucket []byte

	// config for RR integration
	cfg *Config

	// logger
	log logger.Logger

	// gc contains key which are contain timeouts
	gc *sync.Map
	// default timeout for cache cleanup is 1 minute
	timeout time.Duration

	// stop is used to stop keys GC and close boltdb connection
	stop chan struct{}
}

func (s *Plugin) Init(log logger.Logger, cfg config.Configurer) error {
	const op = errors.Op("boltdb_plugin_init")

	if !cfg.Has(PluginName) {
		return errors.E(op, errors.Disabled)
	}

	err := cfg.UnmarshalKey(PluginName, &s.cfg)
	if err != nil {
		return errors.E(op, errors.Disabled, err)
	}

	// add default values
	s.cfg.InitDefaults()

	// set the logger
	s.log = log

	db, err := bolt.Open(path.Join(s.cfg.Dir, s.cfg.File), os.FileMode(s.cfg.Permissions), nil)
	if err != nil {
		return errors.E(op, err)
	}

	// create bucket if it does not exist
	// tx.Commit invokes via the db.Update
	err = db.Update(func(tx *bolt.Tx) error {
		const upOp = errors.Op("boltdb_plugin_update")
		_, err = tx.CreateBucketIfNotExists([]byte(s.cfg.Bucket))
		if err != nil {
			return errors.E(op, upOp)
		}
		return nil
	})

	if err != nil {
		return errors.E(op, err)
	}

	s.DB = db
	s.bucket = []byte(s.cfg.Bucket)
	s.stop = make(chan struct{})
	s.timeout = time.Duration(s.cfg.Interval) * time.Second
	s.gc = &sync.Map{}

	return nil
}

func (s *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	// start the TTL gc
	go s.gcPhase()

	return errCh
}

func (s *Plugin) Stop() error {
	const op = errors.Op("boltdb_plugin_stop")
	err := s.Close()
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

func (s *Plugin) Has(keys ...string) (map[string]bool, error) {
	const op = errors.Op("boltdb_plugin_has")
	s.log.Debug("boltdb HAS method called", "args", keys)
	if keys == nil {
		return nil, errors.E(op, errors.NoKeys)
	}

	m := make(map[string]bool, len(keys))

	// this is readable transaction
	err := s.DB.View(func(tx *bolt.Tx) error {
		// Get retrieves the value for a key in the bucket.
		// Returns a nil value if the key does not exist or if the key is a nested bucket.
		// The returned value is only valid for the life of the transaction.
		for i := range keys {
			keyTrimmed := strings.TrimSpace(keys[i])
			if keyTrimmed == "" {
				return errors.E(op, errors.EmptyKey)
			}
			b := tx.Bucket(s.bucket)
			if b == nil {
				return errors.E(op, errors.NoSuchBucket)
			}
			exist := b.Get([]byte(keys[i]))
			if exist != nil {
				m[keys[i]] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	s.log.Debug("boltdb HAS method finished")
	return m, nil
}

// Get retrieves the value for a key in the bucket.
// Returns a nil value if the key does not exist or if the key is a nested bucket.
// The returned value is only valid for the life of the transaction.
func (s *Plugin) Get(key string) ([]byte, error) {
	const op = errors.Op("boltdb_plugin_get")
	// to get cases like "  "
	keyTrimmed := strings.TrimSpace(key)
	if keyTrimmed == "" {
		return nil, errors.E(op, errors.EmptyKey)
	}

	var val []byte
	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return errors.E(op, errors.NoSuchBucket)
		}
		val = b.Get([]byte(key))

		// try to decode values
		if val != nil {
			buf := bytes.NewReader(val)
			decoder := gob.NewDecoder(buf)

			var i string
			err := decoder.Decode(&i)
			if err != nil {
				// unsafe (w/o runes) convert
				return errors.E(op, err)
			}

			// set the value
			val = []byte(i)
		}
		return nil
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return val, nil
}

func (s *Plugin) MGet(keys ...string) (map[string]interface{}, error) {
	const op = errors.Op("boltdb_plugin_mget")
	// defense
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

	err := s.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucket)
		if b == nil {
			return errors.E(op, errors.NoSuchBucket)
		}

		buf := new(bytes.Buffer)
		var out string
		buf.Grow(100)
		for i := range keys {
			value := b.Get([]byte(keys[i]))
			buf.Write(value)
			// allocate enough space
			dec := gob.NewDecoder(buf)
			if value != nil {
				err := dec.Decode(&out)
				if err != nil {
					return errors.E(op, err)
				}
				m[keys[i]] = out
				buf.Reset()
				out = ""
			}
		}

		return nil
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	return m, nil
}

// Set puts the K/V to the bolt
func (s *Plugin) Set(items ...kv.Item) error {
	const op = errors.Op("boltdb_plugin_set")
	if items == nil {
		return errors.E(op, errors.NoKeys)
	}

	// start writable transaction
	tx, err := s.DB.Begin(true)
	if err != nil {
		return errors.E(op, err)
	}
	defer func() {
		err = tx.Commit()
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				s.log.Error("during the commit, Rollback error occurred", "commit error", err, "rollback error", errRb)
			}
		}
	}()

	b := tx.Bucket(s.bucket)
	// use access by index to avoid copying
	for i := range items {
		// performance note: pass a prepared bytes slice with initial cap
		// we can't move buf and gob out of loop, because we need to clear both from data
		// but gob will contain (w/o re-init) the past data
		buf := bytes.Buffer{}
		encoder := gob.NewEncoder(&buf)
		if errors.Is(errors.EmptyItem, err) {
			return errors.E(op, errors.EmptyItem)
		}

		// Encode value
		err = encoder.Encode(&items[i].Value)
		if err != nil {
			return errors.E(op, err)
		}
		// buf.Bytes will copy the underlying slice. Take a look in case of performance problems
		err = b.Put([]byte(items[i].Key), buf.Bytes())
		if err != nil {
			return errors.E(op, err)
		}

		// if there are no errors, and TTL > 0,  we put the key with timeout to the hashmap, for future check
		// we do not need mutex here, since we use sync.Map
		if items[i].TTL != "" {
			// check correctness of provided TTL
			_, err := time.Parse(time.RFC3339, items[i].TTL)
			if err != nil {
				return errors.E(op, err)
			}
			// Store key TTL in the separate map
			s.gc.Store(items[i].Key, items[i].TTL)
		}

		buf.Reset()
	}

	return nil
}

// Delete all keys from DB
func (s *Plugin) Delete(keys ...string) error {
	const op = errors.Op("boltdb_plugin_delete")
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

	// start writable transaction
	tx, err := s.DB.Begin(true)
	if err != nil {
		return errors.E(op, err)
	}

	defer func() {
		err = tx.Commit()
		if err != nil {
			errRb := tx.Rollback()
			if errRb != nil {
				s.log.Error("during the commit, Rollback error occurred", "commit error", err, "rollback error", errRb)
			}
		}
	}()

	b := tx.Bucket(s.bucket)
	if b == nil {
		return errors.E(op, errors.NoSuchBucket)
	}

	for _, key := range keys {
		err = b.Delete([]byte(key))
		if err != nil {
			return errors.E(op, err)
		}
	}

	return nil
}

// MExpire sets the expiration time to the key
// If key already has the expiration time, it will be overwritten
func (s *Plugin) MExpire(items ...kv.Item) error {
	const op = errors.Op("boltdb_plugin_mexpire")
	for i := range items {
		if items[i].TTL == "" || strings.TrimSpace(items[i].Key) == "" {
			return errors.E(op, errors.Str("should set timeout and at least one key"))
		}

		// verify provided TTL
		_, err := time.Parse(time.RFC3339, items[i].TTL)
		if err != nil {
			return errors.E(op, err)
		}

		s.gc.Store(items[i].Key, items[i].TTL)
	}
	return nil
}

func (s *Plugin) TTL(keys ...string) (map[string]interface{}, error) {
	const op = errors.Op("boltdb_plugin_ttl")
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
		if item, ok := s.gc.Load(keys[i]); ok {
			// a little bit dangerous operation, but user can't store value other that kv.Item.TTL --> int64
			m[keys[i]] = item.(string)
		}
	}
	return m, nil
}

// Close the DB connection
func (s *Plugin) Close() error {
	// stop the keys GC
	s.stop <- struct{}{}
	return s.DB.Close()
}

// RPCService returns associated rpc service.
func (s *Plugin) RPC() interface{} {
	return kv.NewRPCServer(s, s.log)
}

// Name returns plugin name
func (s *Plugin) Name() string {
	return PluginName
}

// ========================= PRIVATE =================================

func (s *Plugin) gcPhase() {
	t := time.NewTicker(s.timeout)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			// calculate current time before loop started to be fair
			now := time.Now()
			s.gc.Range(func(key, value interface{}) bool {
				const op = errors.Op("boltdb_plugin_gc")
				k := key.(string)
				v, err := time.Parse(time.RFC3339, value.(string))
				if err != nil {
					return false
				}

				if now.After(v) {
					// time expired
					s.gc.Delete(k)
					s.log.Debug("key deleted", "key", k)
					err := s.DB.Update(func(tx *bolt.Tx) error {
						b := tx.Bucket(s.bucket)
						if b == nil {
							return errors.E(op, errors.NoSuchBucket)
						}
						err := b.Delete([]byte(k))
						if err != nil {
							return errors.E(op, err)
						}
						return nil
					})
					if err != nil {
						s.log.Error("error during the gc phase of update", "error", err)
						// todo this error is ignored, it means, that timer still be active
						// to prevent this, we need to invoke t.Stop()
						return false
					}
				}
				return true
			})
		case <-s.stop:
			return
		}
	}
}
