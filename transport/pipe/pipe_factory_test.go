package pipe

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/events"
	"github.com/spiral/roadrunner/v2/payload"
	"github.com/spiral/roadrunner/v2/worker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_GetState(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
		assert.Equal(t, worker.StateStopped, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, worker.StateReady, w.State().Value())
	err = w.Stop()
	if err != nil {
		t.Errorf("error stopping the Process: error %v", err)
	}
}

func Test_Kill(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Error(t, w.Wait())
		assert.Equal(t, worker.StateErrored, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, worker.StateReady, w.State().Value())
	err = w.Kill()
	if err != nil {
		t.Errorf("error killing the Process: error %v", err)
	}
	wg.Wait()
}

func Test_Pipe_Start(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	assert.NoError(t, w.Stop())
}

func Test_Pipe_StartError(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	err := cmd.Start()
	if err != nil {
		t.Errorf("error running the command: error %v", err)
	}

	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	_, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("error creating the STDIN pipe: error %v", err)
	}

	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError2(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	// error cause
	_, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("error creating the STDIN pipe: error %v", err)
	}

	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Failboot(t *testing.T) {
	cmd := exec.Command("php", "../../tests/failboot.php")
	ctx := context.Background()

	eb, id := events.Bus()
	defer eb.Unsubscribe(id)
	ch := make(chan events.Event, 10)
	err := eb.SubscribeP(id, "worker.EventWorkerStderr", ch)
	require.NoError(t, err)

	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)

	assert.Nil(t, w)
	assert.Error(t, err)

	ev := <-ch
	if !strings.Contains(ev.Message(), "failboot") {
		t.Fatal("should contain failboot string")
	}
}

func Test_Pipe_Invalid(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("php", "../../tests/invalid.php")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Echo(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Pipe_Echo_Script(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("sh", "../../tests/pipes_test_script.sh")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Pipe_Broken(t *testing.T) {
	t.Parallel()
	cmd := exec.Command("php", "../../tests/client.php", "broken", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		time.Sleep(time.Second)
		err = w.Stop()
		assert.Error(t, err)
	}()

	sw := worker.From(w)

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res)
}

func Benchmark_Pipe_SpawnWorker_Stop(b *testing.B) {
	f := NewPipeFactory()
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
		w, _ := f.SpawnWorkerWithTimeout(context.Background(), cmd)
		go func() {
			if w.Wait() != nil {
				b.Fail()
			}
		}()

		err := w.Stop()
		if err != nil {
			b.Errorf("error stopping the worker: error %v", err)
		}
	}
}

func Benchmark_Pipe_Worker_ExecEcho(b *testing.B) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithTimeout(context.Background(), cmd)
	sw := worker.From(w)

	b.ReportAllocs()
	b.ResetTimer()
	go func() {
		err := w.Wait()
		if err != nil {
			b.Errorf("error waiting the worker: error %v", err)
		}
	}()
	defer func() {
		err := w.Stop()
		if err != nil {
			b.Errorf("error stopping the worker: error %v", err)
		}
	}()

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(&payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEcho3(b *testing.B) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		err = w.Stop()
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(&payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEchoWithoutContext(b *testing.B) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		err = w.Stop()
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(&payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Test_Echo(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	sw := worker.From(w)
	go func() {
		assert.NoError(t, sw.Wait())
	}()
	defer func() {
		err = sw.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_BadPayload(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)

	sw := worker.From(w)

	go func() {
		assert.NoError(t, sw.Wait())
	}()
	defer func() {
		err := sw.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	res, err := sw.Exec(&payload.Payload{})

	assert.Error(t, err)
	assert.Nil(t, res)

	assert.Contains(t, err.Error(), "payload can not be empty")
}

func Test_String(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	assert.Contains(t, w.String(), "php ../../tests/client.php echo pipes")
	assert.Contains(t, w.String(), "ready")
	assert.Contains(t, w.String(), "num_execs: 0")
}

func Test_Echo_Slow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/slow-client.php", "echo", "pipes", "10", "10")

	w, _ := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Broken(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "broken", "pipes")

	eb, id := events.Bus()
	defer eb.Unsubscribe(id)
	ch := make(chan events.Event, 10)
	err := eb.SubscribeP(id, "worker.EventWorkerStderr", ch)
	require.NoError(t, err)

	w, err := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	sw := worker.From(w)

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	time.Sleep(time.Second * 3)
	msg := <-ch
	if strings.ContainsAny(msg.Message(), "undefined_function()") == false {
		t.Fail()
	}
	assert.Error(t, w.Stop())
}

func Test_Error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "error", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()

	defer func() {
		err := w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	res, err := sw.Exec(&payload.Payload{Body: []byte("hello")})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	if errors.Is(errors.SoftJob, err) == false {
		t.Fatal("error should be of type errors.ErrSoftJob")
	}
	assert.Contains(t, err.Error(), "hello")
}

func Test_NumExecs(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithTimeout(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw := worker.From(w)

	_, err := sw.Exec(&payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, uint64(1), w.State().NumExecs())

	_, err = sw.Exec(&payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, uint64(2), w.State().NumExecs())

	_, err = sw.Exec(&payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, uint64(3), w.State().NumExecs())
}
