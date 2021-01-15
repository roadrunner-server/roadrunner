package worker_watcher //nolint:golint,stylecheck
import (
	"context"
	"os/exec"
	"testing"

	"github.com/spiral/roadrunner/v2/interfaces/worker"
	workerImpl "github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func TestNewWorkersStack(t *testing.T) {
	stack := NewWorkersStack()
	assert.Equal(t, int64(0), stack.actualNumOfWorkers)
	assert.Equal(t, []worker.BaseProcess{}, stack.workers)
}

func TestStack_Push(t *testing.T) {
	stack := NewWorkersStack()

	w, err := workerImpl.InitBaseWorker(&exec.Cmd{})
	assert.NoError(t, err)

	stack.Push(w)
	assert.Equal(t, int64(1), stack.actualNumOfWorkers)
}

func TestStack_Pop(t *testing.T) {
	stack := NewWorkersStack()
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")

	w, err := workerImpl.InitBaseWorker(cmd)
	assert.NoError(t, err)

	stack.Push(w)
	assert.Equal(t, int64(1), stack.actualNumOfWorkers)

	_, _ = stack.Pop()
	assert.Equal(t, int64(0), stack.actualNumOfWorkers)
}

func TestStack_FindAndRemoveByPid(t *testing.T) {
	stack := NewWorkersStack()
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := workerImpl.InitBaseWorker(cmd)
	assert.NoError(t, err)

	assert.NoError(t, w.Start())

	stack.Push(w)
	assert.Equal(t, int64(1), stack.actualNumOfWorkers)

	stack.FindAndRemoveByPid(w.Pid())
	assert.Equal(t, int64(0), stack.actualNumOfWorkers)
}

func TestStack_IsEmpty(t *testing.T) {
	stack := NewWorkersStack()
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")

	w, err := workerImpl.InitBaseWorker(cmd)
	assert.NoError(t, err)

	stack.Push(w)
	assert.Equal(t, int64(1), stack.actualNumOfWorkers)

	assert.Equal(t, false, stack.IsEmpty())
}

func TestStack_Workers(t *testing.T) {
	stack := NewWorkersStack()
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := workerImpl.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	stack.Push(w)

	wrks := stack.Workers()
	assert.Equal(t, 1, len(wrks))
	assert.Equal(t, w.Pid(), wrks[0].Pid())
}

func TestStack_Reset(t *testing.T) {
	stack := NewWorkersStack()
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := workerImpl.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	stack.Push(w)
	assert.Equal(t, int64(1), stack.actualNumOfWorkers)
	stack.Reset()
	assert.Equal(t, int64(0), stack.actualNumOfWorkers)
}

func TestStack_Destroy(t *testing.T) {
	stack := NewWorkersStack()
	cmd := exec.Command("php", "../tests/client.php", "echo", "pipes")
	w, err := workerImpl.InitBaseWorker(cmd)
	assert.NoError(t, err)
	assert.NoError(t, w.Start())

	stack.Push(w)
	stack.Destroy(context.Background())
	assert.Equal(t, int64(0), stack.actualNumOfWorkers)
}
