package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"net"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func TestTcpNotStarted(t *testing.T) {
	ls, err := net.Listen("tcp", "localhost:9007")
	defer ls.Close()

	assert.NotNil(t, ls)
	assert.Nil(t, err)

	cmd := exec.Command("php", "tests/client.php", "echo", "")
	assert.Nil(t, cmd.Start())

	w, err := NewSocketFactory(ls, time.Second).NewWorker(cmd)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	assert.Equal(t, "can't attach to running process", err.Error())
}

func TestTcpErrored(t *testing.T) {
	ls, _ := net.Listen("tcp", "localhost:9007")
	defer ls.Close()

	cmd := exec.Command("php", "tests/invalid.php")
	w, err := NewSocketFactory(ls, time.Second).NewWorker(cmd)
	assert.Nil(t, w)
	assert.NotNil(t, err)

	assert.Equal(t, "unable to connect to worker: relay timeout", err.Error())
}

func TestTcpStarted(t *testing.T) {
	ls, _ := net.Listen("tcp", "localhost:9007")
	defer ls.Close()

	cmd := exec.Command("php", "tests/client.php", "echo", "tcp")
	w, err := NewSocketFactory(ls, time.Second).NewWorker(cmd)
	defer w.Stop()

	assert.NotNil(t, w)
	assert.Nil(t, err)
	assert.NotNil(t, *w.Pid)
	assert.Equal(t, cmd.Process.Pid, *w.Pid)
	assert.Equal(t, StateReady, w.st.Value())
}

func TestTcpEcho(t *testing.T) {
	ls, _ := net.Listen("tcp", "localhost:9007")
	defer ls.Close()

	cmd := exec.Command("php", "tests/client.php", "echo", "tcp")
	w, err := NewSocketFactory(ls, time.Second).NewWorker(cmd)

	defer w.Stop()

	r, ctx, err := w.Exec([]byte("hello"), nil)
	assert.Nil(t, err)
	assert.Nil(t, ctx)
	assert.Equal(t, "hello", string(r))
}

func TestUnixEcho(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on windows")
	}

	ls, _ := net.Listen("unix", "sock.unix")
	defer ls.Close()

	cmd := exec.Command("php", "tests/client.php", "echo", "unix")
	w, err := NewSocketFactory(ls, time.Second).NewWorker(cmd)

	defer w.Stop()

	r, ctx, err := w.Exec([]byte("hello"), nil)
	assert.Nil(t, err)
	assert.Nil(t, ctx)
	assert.Equal(t, "hello", string(r))
}
