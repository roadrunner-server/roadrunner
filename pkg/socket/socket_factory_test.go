package socket

import (
	"context"
	"net"
	"os/exec"
	"sync"
	"testing"
	"time"

	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func Test_Tcp_Start(t *testing.T) {
	ctx := context.Background()
	time.Sleep(time.Millisecond * 10) // to ensure free socket

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "tcp")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	err = w.Stop()
	if err != nil {
		t.Errorf("error stopping the Process: error %v", err)
	}
}

func Test_Tcp_StartCloseFactory(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "tcp")

	f := NewSocketServer(ls, time.Minute)
	defer func() {
		err := ls.Close()
		if err != nil {
			t.Errorf("error closing the listener: error %v", err)
		}
	}()

	w, err := f.SpawnWorkerWithTimeout(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	err = w.Stop()
	if err != nil {
		t.Errorf("error stopping the Process: error %v", err)
	}
}

func Test_Tcp_StartError(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	err = cmd.Start()
	if err != nil {
		t.Errorf("error executing the command: error %v", err)
	}

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Tcp_Failboot(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()

	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err3 := ls.Close()
			if err3 != nil {
				t.Errorf("error closing the listener: error %v", err3)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/failboot.php")

	w, err2 := NewSocketServer(ls, time.Second*5).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Nil(t, w)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "failboot")
}

func Test_Tcp_Timeout(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/slow-client.php", "echo", "tcp", "200", "0")

	w, err := NewSocketServer(ls, time.Millisecond*1).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func Test_Tcp_Invalid(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/invalid.php")

	w, err := NewSocketServer(ls, time.Second*1).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Tcp_Broken(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "broken", "tcp")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undefined_function()")
	}()

	defer func() {
		time.Sleep(time.Second)
		err2 := w.Stop()
		// write tcp 127.0.0.1:9007->127.0.0.1:34204: use of closed network connection
		assert.Error(t, err2)
	}()

	sw, err := worker.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})
	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)
	wg.Wait()
}

func Test_Tcp_Echo(t *testing.T) {
	time.Sleep(time.Millisecond * 10) // to ensure free socket
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if assert.NoError(t, err) {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "tcp")

	w, _ := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err = w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := worker.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Unix_Start(t *testing.T) {
	ctx := context.Background()
	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "unix")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	err = w.Stop()
	if err != nil {
		t.Errorf("error stopping the Process: error %v", err)
	}
}

func Test_Unix_Failboot(t *testing.T) {
	ls, err := net.Listen("unix", "sock.unix")
	ctx := context.Background()
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/failboot.php")

	w, err := NewSocketServer(ls, time.Second*5).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failboot")
}

func Test_Unix_Timeout(t *testing.T) {
	ls, err := net.Listen("unix", "sock.unix")
	ctx := context.Background()
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/slow-client.php", "echo", "unix", "200", "0")

	w, err := NewSocketServer(ls, time.Millisecond*100).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func Test_Unix_Invalid(t *testing.T) {
	ctx := context.Background()
	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/invalid.php")

	w, err := NewSocketServer(ls, time.Second*10).SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Unix_Broken(t *testing.T) {
	ctx := context.Background()
	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "broken", "unix")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := w.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "undefined_function()")
	}()

	defer func() {
		time.Sleep(time.Second)
		err = w.Stop()
		assert.Error(t, err)
	}()

	sw, err := worker.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Context)
	assert.Nil(t, res.Body)
	wg.Wait()
}

func Test_Unix_Echo(t *testing.T) {
	ctx := context.Background()
	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				t.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		t.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "unix")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err = w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := worker.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Benchmark_Tcp_SpawnWorker_Stop(b *testing.B) {
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if err == nil {
		defer func() {
			err = ls.Close()
			if err != nil {
				b.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		b.Skip("socket is busy")
	}

	f := NewSocketServer(ls, time.Minute)
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "../../tests/client.php", "echo", "tcp")

		w, err := f.SpawnWorkerWithTimeout(ctx, cmd)
		if err != nil {
			b.Fatal(err)
		}
		go func() {
			assert.NoError(b, w.Wait())
		}()

		err = w.Stop()
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}
}

func Benchmark_Tcp_Worker_ExecEcho(b *testing.B) {
	ctx := context.Background()
	ls, err := net.Listen("tcp", "localhost:9007")
	if err == nil {
		defer func() {
			err = ls.Close()
			if err != nil {
				b.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		b.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "tcp")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		err = w.Stop()
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := worker.From(w)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Unix_SpawnWorker_Stop(b *testing.B) {
	ctx := context.Background()
	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				b.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		b.Skip("socket is busy")
	}

	f := NewSocketServer(ls, time.Minute)
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "../../tests/client.php", "echo", "unix")

		w, err := f.SpawnWorkerWithTimeout(ctx, cmd)
		if err != nil {
			b.Fatal(err)
		}
		err = w.Stop()
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}
}

func Benchmark_Unix_Worker_ExecEcho(b *testing.B) {
	ctx := context.Background()
	ls, err := net.Listen("unix", "sock.unix")
	if err == nil {
		defer func() {
			err := ls.Close()
			if err != nil {
				b.Errorf("error closing the listener: error %v", err)
			}
		}()
	} else {
		b.Skip("socket is busy")
	}

	cmd := exec.Command("php", "../../tests/client.php", "echo", "unix")

	w, err := NewSocketServer(ls, time.Minute).SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		err = w.Stop()
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := worker.From(w)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}
