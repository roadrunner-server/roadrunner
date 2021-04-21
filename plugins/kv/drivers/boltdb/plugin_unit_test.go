package boltdb

import (
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/plugins/kv"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
)

// NewBoltClient instantiate new BOLTDB client
// The parameters are:
// path string 			 -- path to database file (can be placed anywhere), if file is not exist, it will be created
// perm os.FileMode 	 -- file permissions, for example 0777
// options *bolt.Options -- boltDB options, such as timeouts, noGrows options and other
// bucket string 		 -- name of the bucket to use, should be UTF-8
func newBoltClient(path string, perm os.FileMode, options *bolt.Options, bucket string, ttl time.Duration) (kv.Storage, error) {
	const op = errors.Op("boltdb_plugin_new_bolt_client")
	db, err := bolt.Open(path, perm, options)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// bucket should be SET
	if bucket == "" {
		return nil, errors.E(op, errors.Str("bucket should be set"))
	}

	// create bucket if it does not exist
	// tx.Commit invokes via the db.Update
	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return errors.E(op, err)
		}
		return nil
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	// if TTL is not set, make it default
	if ttl == 0 {
		ttl = time.Minute
	}

	l, _ := zap.NewDevelopment()
	s := &Plugin{
		DB:      db,
		bucket:  []byte(bucket),
		stop:    make(chan struct{}),
		timeout: ttl,
		gc:      &sync.Map{},
		log:     logger.NewZapAdapter(l),
	}

	// start the TTL gc
	go s.gcPhase()

	return s, nil
}

func initStorage() kv.Storage {
	storage, err := newBoltClient("rr.db", 0777, nil, "rr", time.Second)
	if err != nil {
		panic(err)
	}
	return storage
}

func cleanup(t *testing.T, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStorage_Has(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	assert.False(t, v["key"])
}

func TestStorage_Has_Set_Has(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	// no such key
	assert.False(t, v["key"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}, kv.Item{
		Key:   "key2",
		Value: "hello world",
		TTL:   "",
	}))

	v, err = s.Has("key", "key2")
	assert.NoError(t, err)
	// no such key
	assert.True(t, v["key"])
	assert.True(t, v["key2"])
}

func TestConcurrentReadWriteTransactions(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	// no such key
	assert.False(t, v["key"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}, kv.Item{
		Key:   "key2",
		Value: "hello world",
		TTL:   "",
	}))

	v, err = s.Has("key", "key2")
	assert.NoError(t, err)
	// no such key
	assert.True(t, v["key"])
	assert.True(t, v["key2"])

	wg := &sync.WaitGroup{}
	wg.Add(3)

	m := &sync.RWMutex{}
	// concurrently set the keys
	go func(s kv.Storage) {
		defer wg.Done()
		for i := 0; i <= 100; i++ {
			m.Lock()
			// set is writable transaction
			// it should stop readable
			assert.NoError(t, s.Set(kv.Item{
				Key:   "key" + strconv.Itoa(i),
				Value: "hello world" + strconv.Itoa(i),
				TTL:   "",
			}, kv.Item{
				Key:   "key2" + strconv.Itoa(i),
				Value: "hello world" + strconv.Itoa(i),
				TTL:   "",
			}))
			m.Unlock()
		}
	}(s)

	// should be no errors
	go func(s kv.Storage) {
		defer wg.Done()
		for i := 0; i <= 100; i++ {
			m.RLock()
			v, err = s.Has("key")
			assert.NoError(t, err)
			// no such key
			assert.True(t, v["key"])
			m.RUnlock()
		}
	}(s)

	// should be no errors
	go func(s kv.Storage) {
		defer wg.Done()
		for i := 0; i <= 100; i++ {
			m.Lock()
			err = s.Delete("key" + strconv.Itoa(i))
			assert.NoError(t, err)
			m.Unlock()
		}
	}(s)

	wg.Wait()
}

func TestStorage_Has_Set_MGet(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	// no such key
	assert.False(t, v["key"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}, kv.Item{
		Key:   "key2",
		Value: "hello world",
		TTL:   "",
	}))

	v, err = s.Has("key", "key2")
	assert.NoError(t, err)
	// no such key
	assert.True(t, v["key"])
	assert.True(t, v["key2"])

	res, err := s.MGet("key", "key2")
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestStorage_Has_Set_Get(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	// no such key
	assert.False(t, v["key"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}, kv.Item{
		Key:   "key2",
		Value: "hello world2",
		TTL:   "",
	}))

	v, err = s.Has("key", "key2")
	assert.NoError(t, err)

	assert.True(t, v["key"])
	assert.True(t, v["key2"])

	res, err := s.Get("key")
	assert.NoError(t, err)

	if string(res) != "hello world" {
		t.Fatal("wrong value by key")
	}
}

func TestStorage_Set_Del_Get(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	// no such key
	assert.False(t, v["key"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}, kv.Item{
		Key:   "key2",
		Value: "hello world",
		TTL:   "",
	}))

	v, err = s.Has("key", "key2")
	assert.NoError(t, err)
	// no such key
	assert.True(t, v["key"])
	assert.True(t, v["key2"])

	// check that keys are present
	res, err := s.MGet("key", "key2")
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	assert.NoError(t, s.Delete("key", "key2"))
	// check that keys are not present
	res, err = s.MGet("key", "key2")
	assert.NoError(t, err)
	assert.Len(t, res, 0)
}

func TestStorage_Set_GetM(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	v, err := s.Has("key")
	assert.NoError(t, err)
	assert.False(t, v["key"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}, kv.Item{
		Key:   "key2",
		Value: "hello world",
		TTL:   "",
	}))

	res, err := s.MGet("key", "key2")
	assert.NoError(t, err)
	assert.Len(t, res, 2)
}

func TestNilAndWrongArgs(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	// check
	v, err := s.Has("key")
	assert.NoError(t, err)
	assert.False(t, v["key"])

	_, err = s.Has("")
	assert.Error(t, err)

	_, err = s.Get("")
	assert.Error(t, err)

	_, err = s.Get(" ")
	assert.Error(t, err)

	_, err = s.Get("                 ")
	assert.Error(t, err)

	_, err = s.MGet("key", "key2", "")
	assert.Error(t, err)

	_, err = s.MGet("key", "key2", "   ")
	assert.Error(t, err)

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	}))

	assert.Error(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "asdf",
	}))

	_, err = s.Has("key")
	assert.NoError(t, err)

	assert.Error(t, s.Set(kv.Item{}))

	err = s.Delete("")
	assert.Error(t, err)

	err = s.Delete("key", "")
	assert.Error(t, err)

	err = s.Delete("key", "     ")
	assert.Error(t, err)

	err = s.Delete("key")
	assert.NoError(t, err)
}

func TestStorage_MExpire_TTL(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	// ensure that storage is clean
	v, err := s.Has("key", "key2")
	assert.NoError(t, err)
	assert.False(t, v["key"])
	assert.False(t, v["key2"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	},
		kv.Item{
			Key:   "key2",
			Value: "hello world",
			TTL:   "",
		}))
	// set timeout to 5 sec
	nowPlusFive := time.Now().Add(time.Second * 5).Format(time.RFC3339)

	i1 := kv.Item{
		Key:   "key",
		Value: "",
		TTL:   nowPlusFive,
	}
	i2 := kv.Item{
		Key:   "key2",
		Value: "",
		TTL:   nowPlusFive,
	}
	assert.NoError(t, s.MExpire(i1, i2))

	time.Sleep(time.Second * 7)

	// ensure that storage is clean
	v, err = s.Has("key", "key2")
	assert.NoError(t, err)
	assert.False(t, v["key"])
	assert.False(t, v["key2"])
}

func TestStorage_SetExpire_TTL(t *testing.T) {
	s := initStorage()
	defer func() {
		if err := s.Close(); err != nil {
			panic(err)
		}
		cleanup(t, "rr.db")
	}()

	// ensure that storage is clean
	v, err := s.Has("key", "key2")
	assert.NoError(t, err)
	assert.False(t, v["key"])
	assert.False(t, v["key2"])

	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "hello world",
		TTL:   "",
	},
		kv.Item{
			Key:   "key2",
			Value: "hello world",
			TTL:   "",
		}))

	nowPlusFive := time.Now().Add(time.Second * 5).Format(time.RFC3339)

	// set timeout to 5 sec
	assert.NoError(t, s.Set(kv.Item{
		Key:   "key",
		Value: "value",
		TTL:   nowPlusFive,
	},
		kv.Item{
			Key:   "key2",
			Value: "value",
			TTL:   nowPlusFive,
		}))

	time.Sleep(time.Second * 2)
	m, err := s.TTL("key", "key2")
	assert.NoError(t, err)

	// remove a precision 4.02342342 -> 4
	keyTTL, err := strconv.Atoi(m["key"].(string)[0:1])
	if err != nil {
		t.Fatal(err)
	}

	// remove a precision 4.02342342 -> 4
	key2TTL, err := strconv.Atoi(m["key"].(string)[0:1])
	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, keyTTL < 5)
	assert.True(t, key2TTL < 5)

	time.Sleep(time.Second * 7)

	// ensure that storage is clean
	v, err = s.Has("key", "key2")
	assert.NoError(t, err)
	assert.False(t, v["key"])
	assert.False(t, v["key2"])
}
