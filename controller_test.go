package roadrunner

import (
	"fmt"
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

func (w *eWatcher) Attach(p Pool) Controller {
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

func (w *eWatcher) remove(wr *Worker, err error) {
	w.p.Remove(wr, err)
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

	rr.Attach(&eWatcher{})
	assert.NoError(t, rr.Start())

	assert.NotNil(t, rr.pController)
	assert.Equal(t, rr.pController.(*eWatcher).p, rr.pool)

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

	rr.Attach(&eWatcher{})
	assert.NoError(t, rr.Start())

	assert.NotNil(t, rr.pController)
	assert.Equal(t, rr.pController.(*eWatcher).p, rr.pool)

	oldWatcher := rr.pController

	assert.NoError(t, rr.Reset())

	assert.NotNil(t, rr.pController)
	assert.Equal(t, rr.pController.(*eWatcher).p, rr.pool)
	assert.NotEqual(t, oldWatcher, rr.pController)

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

	rr.Attach(&eWatcher{
		onAttach: func(p Pool) {
			attachedPool = p
		},
		onDetach: func(p Pool) {
			assert.Equal(t, attachedPool, p)
		},
	})
	assert.NoError(t, rr.Start())

	assert.NotNil(t, rr.pController)
	assert.Equal(t, rr.pController.(*eWatcher).p, rr.pool)

	res, err := rr.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_RemoveWorkerOnAllocation(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php pid pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	rr.Attach(&eWatcher{})
	assert.NoError(t, rr.Start())

	wr := rr.Workers()[0]

	res, err := rr.Exec(&Payload{Body: []byte("hello")})
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%v", *wr.Pid), res.String())
	lastPid := res.String()

	rr.pController.(*eWatcher).remove(wr, nil)

	res, err = rr.Exec(&Payload{Body: []byte("hello")})
	assert.NoError(t, err)
	assert.NotEqual(t, lastPid, res.String())

	assert.NotEqual(t, StateReady, wr.state.Value())

	_, ok := rr.pool.(*StaticPool).remove.Load(wr)
	assert.False(t, ok)
}

func Test_RemoveWorkerAfterTask(t *testing.T) {
	rr := NewServer(
		&ServerConfig{
			Command: "php tests/client.php slow-pid pipes",
			Relay:   "pipes",
			Pool: &Config{
				NumWorkers:      1,
				AllocateTimeout: time.Second,
				DestroyTimeout:  time.Second,
			},
		})
	defer rr.Stop()

	rr.Attach(&eWatcher{})
	assert.NoError(t, rr.Start())

	wr := rr.Workers()[0]
	lastPid := ""

	wait := make(chan interface{})
	go func() {
		res, err := rr.Exec(&Payload{Body: []byte("hello")})
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("%v", *wr.Pid), res.String())
		lastPid = res.String()

		close(wait)
	}()

	// wait for worker execution to be in progress
	time.Sleep(time.Millisecond * 250)
	rr.pController.(*eWatcher).remove(wr, nil)

	<-wait

	// must be replaced
	assert.NotEqual(t, lastPid, fmt.Sprintf("%v", rr.Workers()[0]))

	// must not be registered within the pool
	rr.pController.(*eWatcher).remove(wr, nil)
}
