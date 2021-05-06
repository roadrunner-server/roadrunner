package pool

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/transport/pipe"
	"github.com/stretchr/testify/assert"
)

var cfgSupervised = Config{
	NumWorkers:      uint64(1),
	AllocateTimeout: time.Second,
	DestroyTimeout:  time.Second,
	Supervisor: &SupervisorConfig{
		WatchTick:       1 * time.Second,
		TTL:             100 * time.Second,
		IdleTTL:         100 * time.Second,
		ExecTTL:         100 * time.Second,
		MaxWorkerMemory: 100,
	},
}

func TestSupervisedPool_Exec(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/memleak.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgSupervised,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	pidBefore := p.Workers()[0].Pid()

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		_, err = p.Exec(payload.Payload{
			Context: []byte(""),
			Body:    []byte("foo"),
		})
		assert.NoError(t, err)
	}

	assert.NotEqual(t, pidBefore, p.Workers()[0].Pid())

	p.Destroy(context.Background())
}

// This test should finish without freezes
func TestSupervisedPool_ExecWithDebugMode(t *testing.T) {
	var cfgSupervised = cfgSupervised
	cfgSupervised.Debug = true

	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/memleak.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgSupervised,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 100)
		_, err = p.Exec(payload.Payload{
			Context: []byte(""),
			Body:    []byte("foo"),
		})
		assert.NoError(t, err)
	}

	p.Destroy(context.Background())
}

func TestSupervisedPool_ExecTTL_TimedOut(t *testing.T) {
	var cfgExecTTL = Config{
		NumWorkers:      uint64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1 * time.Second,
			TTL:             100 * time.Second,
			IdleTTL:         100 * time.Second,
			ExecTTL:         1 * time.Second,
			MaxWorkerMemory: 100,
		},
	}
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/sleep.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Destroy(context.Background())

	pid := p.Workers()[0].Pid()

	resp, err := p.execWithTTL(context.Background(), payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.Error(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second * 1)
	// should be new worker with new pid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
}

func TestSupervisedPool_Idle(t *testing.T) {
	var cfgExecTTL = Config{
		NumWorkers:      uint64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1 * time.Second,
			TTL:             100 * time.Second,
			IdleTTL:         1 * time.Second,
			ExecTTL:         100 * time.Second,
			MaxWorkerMemory: 100,
		},
	}
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/idle.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	pid := p.Workers()[0].Pid()

	resp, err := p.execWithTTL(context.Background(), payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.Nil(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second * 5)

	// worker should be marked as invalid and reallocated
	_, err = p.execWithTTL(context.Background(), payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})
	assert.NoError(t, err)
	// should be new worker with new pid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
	p.Destroy(context.Background())
}

func TestSupervisedPool_ExecTTL_OK(t *testing.T) {
	var cfgExecTTL = Config{
		NumWorkers:      uint64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1 * time.Second,
			TTL:             100 * time.Second,
			IdleTTL:         100 * time.Second,
			ExecTTL:         4 * time.Second,
			MaxWorkerMemory: 100,
		},
	}
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/exec_ttl.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Destroy(context.Background())

	pid := p.Workers()[0].Pid()

	time.Sleep(time.Millisecond * 100)
	resp, err := p.Exec(payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second * 1)
	// should be the same pid
	assert.Equal(t, pid, p.Workers()[0].Pid())
}

func TestSupervisedPool_MaxMemoryReached(t *testing.T) {
	var cfgExecTTL = Config{
		NumWorkers:      uint64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1 * time.Second,
			TTL:             100 * time.Second,
			IdleTTL:         100 * time.Second,
			ExecTTL:         4 * time.Second,
			MaxWorkerMemory: 1,
		},
	}

	block := make(chan struct{}, 10)
	listener := func(event interface{}) {
		if ev, ok := event.(events.PoolEvent); ok {
			if ev.Event == events.EventMaxMemory {
				block <- struct{}{}
			}
		}
	}

	// constructed
	// max memory
	// constructed
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/memleak.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
		AddListeners(listener),
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	resp, err := p.Exec(payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	<-block
	p.Destroy(context.Background())
}
