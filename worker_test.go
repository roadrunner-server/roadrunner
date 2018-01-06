package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
	"time"
)

func TestGetState(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	assert.Equal(t, StateReady, w.State().Value())
}

func TestStop(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	w.Stop()
	assert.Equal(t, StateStopped, w.State().Value())
}

func TestEcho(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	r, ctx, err := w.Exec([]byte("hello"), nil)
	assert.Nil(t, err)
	assert.Nil(t, ctx)
	assert.Equal(t, "hello", string(r))
}

func TestError(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "error", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	r, ctx, err := w.Exec([]byte("hello"), nil)
	assert.Nil(t, r)
	assert.NotNil(t, err)
	assert.Nil(t, ctx)

	assert.IsType(t, JobError{}, err)
	assert.Equal(t, "hello", err.Error())
}

func TestBroken(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "broken", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	r, ctx, err := w.Exec([]byte("hello"), nil)
	assert.Nil(t, r)
	assert.NotNil(t, err)
	assert.Nil(t, ctx)

	assert.IsType(t, WorkerError(""), err)
	assert.Contains(t, err.Error(), "undefined_function()")
}

func TestCalled(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	w.Exec([]byte("hello"), nil)
	assert.Equal(t, uint64(1), w.NumExecs())

	w.Exec([]byte("hello"), nil)
	assert.Equal(t, uint64(2), w.NumExecs())

	w.Exec([]byte("hello"), nil)
	assert.Equal(t, uint64(3), w.NumExecs())
}

func TestStateChanged(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	tm := time.Now()
	w.Exec([]byte("hello"), nil)
	assert.True(t, w.State().Updated().After(tm))
}
