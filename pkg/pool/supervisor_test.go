package pool

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/transport/pipe"
	"github.com/spiral/roadrunner/v2/tools"
	"github.com/stretchr/testify/assert"
)

var cfgSupervised = Config{
	NumWorkers:      int64(1),
	AllocateTimeout: time.Second,
	DestroyTimeout:  time.Second,
	Supervisor: &SupervisorConfig{
		WatchTick:       1,
		TTL:             100,
		IdleTTL:         100,
		ExecTTL:         100,
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
	stopCh := make(chan struct{})
	defer p.Destroy(context.Background())

	go func() {
		for {
			select {
			case <-stopCh:
				return
			default:
				workers := p.Workers()
				if len(workers) > 0 {
					s, err := tools.WorkerProcessState(workers[0])
					assert.NoError(t, err)
					assert.NotNil(t, s)
					// since this is soft limit, double max memory limit watch
					if (s.MemoryUsage / MB) > cfgSupervised.Supervisor.MaxWorkerMemory*2 {
						assert.Fail(t, "max memory reached")
					}
				}
			}
		}
	}()

	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * 50)
		_, err = p.Exec(payload.Payload{
			Context: []byte(""),
			Body:    []byte("foo"),
		})
		assert.NoError(t, err)
	}

	stopCh <- struct{}{}
}

func TestSupervisedPool_ExecTTL_TimedOut(t *testing.T) {
	var cfgExecTTL = Config{
		NumWorkers:      int64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1,
			TTL:             100,
			IdleTTL:         100,
			ExecTTL:         1,
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

	resp, err := p.ExecWithContext(context.Background(), payload.Payload{
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
		NumWorkers:      int64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1,
			TTL:             100,
			IdleTTL:         1,
			ExecTTL:         100,
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

	pid := p.Workers()[0].Pid()

	resp, err := p.ExecWithContext(context.Background(), payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.Nil(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second * 5)
	// should be new worker with new pid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
	p.Destroy(context.Background())
}

func TestSupervisedPool_ExecTTL_OK(t *testing.T) {
	var cfgExecTTL = Config{
		NumWorkers:      int64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1,
			TTL:             100,
			IdleTTL:         100,
			ExecTTL:         4,
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
		NumWorkers:      int64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1,
			TTL:             100,
			IdleTTL:         100,
			ExecTTL:         4,
			MaxWorkerMemory: 1,
		},
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

	time.Sleep(time.Second * 2)
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
}
