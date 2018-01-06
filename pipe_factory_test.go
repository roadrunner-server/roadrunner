package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func TestPipeNotStarted(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	assert.Nil(t, cmd.Start())

	w, err := new(PipeFactory).NewWorker(cmd)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	assert.Equal(t, "can't attach to running process", err.Error())
}

func TestPipeErrored(t *testing.T) {
	cmd := exec.Command("php", "tests/invalid.php")
	w, err := new(PipeFactory).NewWorker(cmd)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	assert.Equal(t, "unable to connect to worker: unexpected response, `control` header is missing", err.Error())
}

func TestPipeStarted(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, err := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	assert.NotNil(t, w)
	assert.Nil(t, err)
	assert.NotNil(t, *w.Pid)
	assert.Equal(t, cmd.Process.Pid, *w.Pid)
	assert.Equal(t, StateReady, w.st.Value())
}

func TestPipeEcho(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	w, _ := new(PipeFactory).NewWorker(cmd)
	defer w.Stop()

	r, ctx, err := w.Exec([]byte("hello"), nil)
	assert.Nil(t, err)
	assert.Nil(t, ctx)
	assert.Equal(t, "hello", string(r))
}
