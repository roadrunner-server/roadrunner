package roadrunner

import (
	"net"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_UnixSocket_Start(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "unix", "withpid")

	w, err := NewUnixSocketFactory("sock.unix", time.Minute).SpawnWorker(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	w.Stop()
}

func Test_UnixSocket_Echo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "unix", "withpid")

	w, err := NewUnixSocketFactory("sock.unix", time.Minute).SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Benchmark_UnixSocket_SpawnWorker_Stop(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("not supported on " + runtime.GOOS)
	}

	f := NewUnixSocketFactory("sock.unix", time.Minute)
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "tests/client.php", "echo", "unix", "withpid")

		w, _ := f.SpawnWorker(cmd)
		go func() {
			if w.Wait() != nil {
				b.Fail()
			}
		}()

		w.Stop()
	}
}

func Benchmark_UnixSocket_Worker_ExecEcho(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("not supported on " + runtime.GOOS)
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "unix", "withpid")

	w, _ := NewUnixSocketFactory("sock.unix", time.Minute).SpawnWorker(cmd)
	go func() {
		w.Wait()
	}()
	defer w.Stop()

	for n := 0; n < b.N; n++ {
		if _, err := w.Exec(&Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_UnixSocket_Pool_ExecEcho(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("not supported on " + runtime.GOOS)
	}

	p, _ := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "unix", "withpid") },
		NewUnixSocketFactory("sock.unix", time.Minute),
		Config{
			NumWorkers:      8,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(&Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Unix_Pool_ExecEcho(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		b.Skip("socket is busy")
	}

	p, _ := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "unix") },
		NewSocketFactory(ls, time.Minute),
		Config{
			NumWorkers:      8,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(&Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}
