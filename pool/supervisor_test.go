package pool

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/transport/pipe"
	"github.com/spiral/roadrunner/v2/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var cfgSupervised = &Config{
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
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/memleak.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgSupervised,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	time.Sleep(time.Second)

	pidBefore := p.Workers()[0].Pid()

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		_, err = p.Exec(&payload.Payload{
			Context: []byte(""),
			Body:    []byte("foo"),
		})
		assert.NoError(t, err)
	}

	assert.NotEqual(t, pidBefore, p.Workers()[0].Pid())

	p.Destroy(context.Background())
}

func Test_SupervisedPoolReset(t *testing.T) {
	ctx := context.Background()
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		cfgSupervised,
	)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	w := p.Workers()
	if len(w) == 0 {
		t.Fatal("should be workers inside")
	}

	pid := w[0].Pid()
	require.NoError(t, p.Reset(context.Background()))

	w2 := p.Workers()
	if len(w2) == 0 {
		t.Fatal("should be workers inside")
	}

	require.NotEqual(t, pid, w2[0].Pid())
}

// This test should finish without freezes
func TestSupervisedPool_ExecWithDebugMode(t *testing.T) {
	var cfgSupervised = cfgSupervised
	cfgSupervised.Debug = true

	ctx := context.Background()
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/supervised.php") },
		pipe.NewPipeFactory(),
		cfgSupervised,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Second)
		_, err = p.Exec(&payload.Payload{
			Context: []byte(""),
			Body:    []byte("foo"),
		})
		assert.NoError(t, err)
	}

	p.Destroy(context.Background())
}

func TestSupervisedPool_ExecTTL_TimedOut(t *testing.T) {
	var cfgExecTTL = &Config{
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
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/sleep.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Destroy(context.Background())

	pid := p.Workers()[0].Pid()

	resp, err := p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.Error(t, err)
	assert.Empty(t, resp)

	time.Sleep(time.Second * 1)
	// should be new worker with new pid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
}

func TestSupervisedPool_ExecTTL_WorkerRestarted(t *testing.T) {
	var cfgExecTTL = &Config{
		NumWorkers: uint64(1),
		Supervisor: &SupervisorConfig{
			WatchTick: 1 * time.Second,
			TTL:       5 * time.Second,
		},
	}
	ctx := context.Background()
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/sleep-ttl.php") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	pid := p.Workers()[0].Pid()

	resp, err := p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Equal(t, string(resp.Body), "hello world")
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second)
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
	require.Equal(t, p.Workers()[0].State().Value(), worker.StateReady)
	pid = p.Workers()[0].Pid()

	resp, err = p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Equal(t, string(resp.Body), "hello world")
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second)
	// should be new worker with new pid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
	require.Equal(t, p.Workers()[0].State().Value(), worker.StateReady)

	p.Destroy(context.Background())
}

func TestSupervisedPool_Idle(t *testing.T) {
	var cfgExecTTL = &Config{
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
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/idle.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	pid := p.Workers()[0].Pid()

	resp, err := p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second * 5)

	// worker should be marked as invalid and reallocated
	rsp, err := p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})
	assert.NoError(t, err)
	require.NotNil(t, rsp)
	time.Sleep(time.Second * 2)
	require.Len(t, p.Workers(), 1)
	// should be new worker with new pid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
	p.Destroy(context.Background())
}

func TestSupervisedPool_IdleTTL_StateAfterTimeout(t *testing.T) {
	var cfgExecTTL = &Config{
		NumWorkers:      uint64(1),
		AllocateTimeout: time.Second,
		DestroyTimeout:  time.Second,
		Supervisor: &SupervisorConfig{
			WatchTick:       1 * time.Second,
			TTL:             1 * time.Second,
			IdleTTL:         1 * time.Second,
			MaxWorkerMemory: 100,
		},
	}
	ctx := context.Background()
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/exec_ttl.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Destroy(context.Background())

	pid := p.Workers()[0].Pid()

	time.Sleep(time.Millisecond * 100)
	resp, err := p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second * 2)

	if len(p.Workers()) < 1 {
		t.Fatal("should be at least 1 worker")
		return
	}
	// should be destroyed, state should be Ready, not Invalid
	assert.NotEqual(t, pid, p.Workers()[0].Pid())
	assert.Equal(t, int64(1), p.Workers()[0].State().Value())
}

func TestSupervisedPool_ExecTTL_OK(t *testing.T) {
	var cfgExecTTL = &Config{
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
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/exec_ttl.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)
	defer p.Destroy(context.Background())

	pid := p.Workers()[0].Pid()

	time.Sleep(time.Millisecond * 100)
	resp, err := p.Exec(&payload.Payload{
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
	var cfgExecTTL = &Config{
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

	// constructed
	// max memory
	// constructed
	ctx := context.Background()
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/memleak.php", "pipes") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	resp, err := p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	assert.NoError(t, err)
	assert.Empty(t, resp.Body)
	assert.Empty(t, resp.Context)

	time.Sleep(time.Second)
	p.Destroy(context.Background())
}

func TestSupervisedPool_AllocateFailedOK(t *testing.T) {
	var cfgExecTTL = &Config{
		NumWorkers:      uint64(2),
		AllocateTimeout: time.Second * 15,
		DestroyTimeout:  time.Second * 5,
		Supervisor: &SupervisorConfig{
			WatchTick: 1 * time.Second,
			TTL:       5 * time.Second,
		},
	}

	ctx := context.Background()
	p, err := NewStaticPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../tests/allocate-failed.php") },
		pipe.NewPipeFactory(),
		cfgExecTTL,
	)

	assert.NoError(t, err)
	require.NotNil(t, p)

	time.Sleep(time.Second)

	// should be ok
	_, err = p.Exec(&payload.Payload{
		Context: []byte(""),
		Body:    []byte("foo"),
	})

	require.NoError(t, err)

	// after creating this file, PHP will fail
	file, err := os.Create("break")
	require.NoError(t, err)

	time.Sleep(time.Second * 5)
	assert.NoError(t, file.Close())
	assert.NoError(t, os.Remove("break"))

	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "panic should not be fired!")
		} else {
			p.Destroy(context.Background())
		}
	}()
}
