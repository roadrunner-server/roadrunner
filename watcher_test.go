package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
	"time"
)

type eWatcher struct {
	p        Pool
	onAttach func(p Pool)
	onDetach func(p Pool)
}

func (w *eWatcher) Attach(p Pool) Watcher {
	wp := &eWatcher{p: p, onAttach: w.onAttach, onDetach: w.onDetach}

	if wp.onAttach != nil {
		wp.onAttach(p)
	}

	return wp
}

func (w *eWatcher) Detach() {
	if w.onDetach != nil {
		w.onDetach(w.p)
	}
}

func Test_WatcherWatch(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	rr.Watch(&eWatcher{})
	assert.NoError(t, rr.Start())

	assert.NotNil(t, rr.pWatcher)
	assert.Equal(t, rr.pWatcher.(*eWatcher).p, rr.pool)

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_WatcherReattach(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	rr.Watch(&eWatcher{})
	assert.NoError(t, rr.Start())

	assert.NotNil(t, rr.pWatcher)
	assert.Equal(t, rr.pWatcher.(*eWatcher).p, rr.pool)

	oldWatcher := rr.pWatcher

	assert.NoError(t, rr.Reset())

	assert.NotNil(t, rr.pWatcher)
	assert.Equal(t, rr.pWatcher.(*eWatcher).p, rr.pool)
	assert.NotEqual(t, oldWatcher, rr.pWatcher)

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_WatcherAttachDetachSequence(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php echo pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      int64(runtime.NumCPU()),
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	var attachedPool Pool

	rr.Watch(&eWatcher{
		onAttach: func(p Pool) {
			attachedPool = p
		},
		onDetach: func(p Pool) {
			assert.Equal(t, attachedPool, p)
		},
	})
	assert.NoError(t, rr.Start())

	assert.NotNil(t, rr.pWatcher)
	assert.Equal(t, rr.pWatcher.(*eWatcher).p, rr.pool)

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}
