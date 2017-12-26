package roadrunner

import (
	"github.com/spiral/goridge"
	"github.com/stretchr/testify/assert"
	"io"
	"os/exec"
	"testing"
	"time"
)

func getPipes(cmd *exec.Cmd) (io.ReadCloser, io.WriteCloser) {
	in, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	out, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	return in, out
}

func TestOnStarted(t *testing.T) {
	pr := exec.Command("php", "tests/echo-client.php")
	pr.Start()

	_, err := NewWorker(pr)
	assert.NotNil(t, err)
	assert.Equal(t, "can't attach to running process", err.Error())
}

func TestNewWorkerState(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/echo-client.php"))
	assert.Nil(t, err)
	assert.Equal(t, StateInactive, w.State)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Equal(t, StateBooting, w.State)

	assert.Nil(t, w.Start())
	assert.Equal(t, StateReady, w.State)
}

func TestStop(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/echo-client.php"))
	assert.Nil(t, err)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Nil(t, w.Start())

	w.Stop()
	assert.Equal(t, StateStopped, w.State)
}

func TestEcho(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/echo-client.php"))
	assert.Nil(t, err)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Nil(t, w.Start())

	r, ctx, err := w.Execute([]byte("hello"), nil)
	assert.Nil(t, err)
	assert.Nil(t, ctx)
	assert.Equal(t, "hello", string(r))
}

func TestError(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/error-client.php"))
	assert.Nil(t, err)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Nil(t, w.Start())

	r, ctx, err := w.Execute([]byte("hello"), nil)
	assert.Nil(t, r)
	assert.NotNil(t, err)
	assert.Nil(t, ctx)

	assert.IsType(t, JobError{}, err)
	assert.Equal(t, "hello", err.Error())
}

func TestBroken(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/broken-client.php"))
	assert.Nil(t, err)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Nil(t, w.Start())

	r, ctx, err := w.Execute([]byte("hello"), nil)
	assert.Nil(t, r)
	assert.NotNil(t, err)
	assert.Nil(t, ctx)

	assert.IsType(t, WorkerError(""), err)
	assert.Contains(t, err.Error(), "undefined_function()")
}

func TestNumExecutions(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/echo-client.php"))
	assert.Nil(t, err)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Nil(t, w.Start())

	w.Execute([]byte("hello"), nil)
	assert.Equal(t, uint64(1), w.NumExecutions)

	w.Execute([]byte("hello"), nil)
	assert.Equal(t, uint64(2), w.NumExecutions)

	w.Execute([]byte("hello"), nil)
	assert.Equal(t, uint64(3), w.NumExecutions)
}

func TestLastExecution(t *testing.T) {
	w, err := NewWorker(exec.Command("php", "tests/echo-client.php"))
	assert.Nil(t, err)

	w.attach(goridge.NewPipeRelay(getPipes(w.cmd)))
	assert.Nil(t, w.Start())

	tm := time.Now()
	w.Execute([]byte("hello"), nil)
	assert.True(t, w.Last.After(tm))
}
