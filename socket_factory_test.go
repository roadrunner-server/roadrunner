package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"net"
	"os/exec"
	"runtime"
	"testing"
	"time"
)

func Test_Tcp_Start(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "tcp")

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	w.Stop()
}

func Test_Tcp_StartCloseFactory(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "tcp")

	f := NewSocketFactory(ls, time.Minute)
	defer f.Close()

	w, err := f.SpawnWorker(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	w.Stop()
}

func Test_Tcp_StartError(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	cmd.Start()

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Tcp_Failboot(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/failboot.php")

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failboot")
}

func Test_Tcp_Timeout(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/slow-client.php", "echo", "tcp", "200", "0")

	w, err := NewSocketFactory(ls, time.Millisecond*100).SpawnWorker(cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "relay timeout")
}

func Test_Tcp_Invalid(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/invalid.php")

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Tcp_Broken(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "broken", "tcp")

	w, _ := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	go func() {
		err := w.Wait()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undefined_function()")
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res)
}

func Test_Tcp_Echo(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "tcp")

	w, _ := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
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

func Test_Unix_Start(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "unix")

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	w.Stop()
}

func Test_Unix_Failboot(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/failboot.php")

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failboot")
}

func Test_Unix_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/slow-client.php", "echo", "unix", "200", "0")

	w, err := NewSocketFactory(ls, time.Millisecond*100).SpawnWorker(cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "relay timeout")
}

func Test_Unix_Invalid(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/invalid.php")

	w, err := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Unix_Broken(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "broken", "unix")

	w, _ := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
	go func() {
		err := w.Wait()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undefined_function()")
	}()
	defer w.Stop()

	res, err := w.Exec(&Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res)
}

func Test_Unix_Echo(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "unix")

	w, _ := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
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

func Benchmark_Tcp_SpawnWorker_Stop(b *testing.B) {
	ls, err := net.Listen("tcp", "localhost:9007")
	if err == nil {
		defer ls.Close()
	} else {
		b.Skip("socket is busy")
	}

	f := NewSocketFactory(ls, time.Minute)
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "tests/client.php", "echo", "tcp")

		w, _ := f.SpawnWorker(cmd)
		go func() {
			if w.Wait() != nil {
				b.Fail()
			}
		}()

		w.Stop()
	}
}

func Benchmark_Tcp_Worker_ExecEcho(b *testing.B) {
	ls, err := net.Listen("tcp", "localhost:9007")
	if err == nil {
		defer ls.Close()
	} else {
		b.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "tcp")

	w, _ := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
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

func Benchmark_Unix_SpawnWorker_Stop(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		b.Skip("socket is busy")
	}

	f := NewSocketFactory(ls, time.Minute)
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "tests/client.php", "echo", "unix")

		w, _ := f.SpawnWorker(cmd)
		go func() {
			if w.Wait() != nil {
				b.Fail()
			}
		}()

		w.Stop()
	}
}

func Benchmark_Unix_Worker_ExecEcho(b *testing.B) {
	if runtime.GOOS == "windows" {
		b.Skip("not supported on " + runtime.GOOS)
	}

	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer ls.Close()
	} else {
		b.Skip("socket is busy")
	}

	cmd := exec.Command("php", "tests/client.php", "echo", "unix")

	w, _ := NewSocketFactory(ls, time.Minute).SpawnWorker(cmd)
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
