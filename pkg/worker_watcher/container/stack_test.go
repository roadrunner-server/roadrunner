package container

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestNewWorkersStack(t *testing.T) {
	stack := NewWorkersStack(0)
	assert.Equal(t, uint64(0), stack.actualNumOfWorkers)
	assert.Equal(t, []worker.BaseProcess{}, stack.workers)
}

func TestStack_Push(t *testing.T) {
	stack := NewWorkersStack(1)

	w, err := worker.InitBaseWorker(&exec.Cmd{})
	assert.NoError(t, err)

	sw := worker.From(w)

	stack.Push(sw)
	assert.Equal(t, uint64(1), stack.actualNumOfWorkers)
}

func TestStack_Pop(t *testing.T) {
	stack := NewWorkersStack(1)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")

	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)

	sw := worker.From(w)

	stack.Push(sw)
	assert.Equal(t, uint64(1), stack.actualNumOfWorkers)

	_, _ = stack.Pop()
	assert.Equal(t, uint64(0), stack.actualNumOfWorkers)
}

func TestStack_FindAndRemoveByPid(t *testing.T) {
	stack := NewWorkersStack(1)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)

	assert.NoError(t, w.Start())

	sw := worker.From(w)

	stack.Push(sw)
	assert.Equal(t, uint64(1), stack.actualNumOfWorkers)

	stack.FindAndRemoveByPid(w.Pid())
	assert.Equal(t, uint64(0), stack.actualNumOfWorkers)
}

func TestStack_IsEmpty(t *testing.T) {
	stack := NewWorkersStack(1)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")

	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)

	sw := worker.From(w)
	stack.Push(sw)

	assert.Equal(t, uint64(1), stack.actualNumOfWorkers)

	assert.Equal(t, false, stack.IsEmpty())
}

func TestStack_Workers(t *testing.T) {
	stack := NewWorkersStack(1)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	sw := worker.From(w)
	stack.Push(sw)

	wrks := stack.Workers()
	assert.Equal(t, 1, len(wrks))
	assert.Equal(t, w.Pid(), wrks[0].Pid())
}

func TestStack_Reset(t *testing.T) {
	stack := NewWorkersStack(1)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	sw := worker.From(w)
	stack.Push(sw)

	assert.Equal(t, uint64(1), stack.actualNumOfWorkers)
	stack.Reset()
	assert.Equal(t, uint64(0), stack.actualNumOfWorkers)
}

func TestStack_Destroy(t *testing.T) {
	stack := NewWorkersStack(1)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	sw := worker.From(w)
	stack.Push(sw)

	stack.Destroy(context.Background())
	assert.Equal(t, uint64(0), stack.actualNumOfWorkers)
}

func TestStack_DestroyWithWait(t *testing.T) {
	stack := NewWorkersStack(2)
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := worker.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	sw := worker.From(w)
	stack.Push(sw)
	stack.Push(sw)
	assert.Equal(t, uint64(2), stack.actualNumOfWorkers)

	go func() {
		wrk, _ := stack.Pop()
		time.Sleep(time.Second * 3)
		stack.Push(wrk)
	}()
	time.Sleep(time.Second)
	stack.Destroy(context.Background())
	assert.Equal(t, uint64(0), stack.actualNumOfWorkers)
}
