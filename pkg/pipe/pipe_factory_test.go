package pipe

import (
	"context"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/interfaces/events"
	"github.com/spiral/roadrunner/v2/internal"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	workerImpl "github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func Test_GetState(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
		assert.Equal(t, internal.StateStopped, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, internal.StateReady, w.State().Value())
	err = w.Stop(ctx)
	if err != nil {
		t.Errorf("error stopping the Process: error %v", err)
	}
}

func Test_Kill(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		assert.Error(t, w.Wait())
		// TODO changed from stopped, discuss
		assert.Equal(t, internal.StateErrored, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, internal.StateReady, w.State().Value())
	err = w.Kill()
	if err != nil {
		t.Errorf("error killing the Process: error %v", err)
	}
	wg.Wait()
}

func Test_Pipe_Start(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	assert.NoError(t, w.Stop(ctx))
}

func Test_Pipe_StartError(t *testing.T) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	err := cmd.Start()
	if err != nil {
		t.Errorf("error running the command: error %v", err)
	}

	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError(t *testing.T) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	_, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("error creating the STDIN pipe: error %v", err)
	}

	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError2(t *testing.T) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	_, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("error creating the STDIN pipe: error %v", err)
	}

	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Failboot(t *testing.T) {
	cmd := exec.Command("php", "../../tests/failboot.php")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)

	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failboot")
}

func Test_Pipe_Invalid(t *testing.T) {
	cmd := exec.Command("php", "../../tests/invalid.php")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Echo(t *testing.T) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = w.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := workerImpl.From(w)
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

func Test_Pipe_Broken(t *testing.T) {
	cmd := exec.Command("php", "../../tests/client.php", "broken", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		time.Sleep(time.Second)
		err = w.Stop(ctx)
		assert.Error(t, err)
	}()

	sw, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)
}

func Benchmark_Pipe_SpawnWorker_Stop(b *testing.B) {
	f := NewPipeFactory()
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
		w, _ := f.SpawnWorkerWithContext(context.Background(), cmd)
		go func() {
			if w.Wait() != nil {
				b.Fail()
			}
		}()

		err := w.Stop(context.Background())
		if err != nil {
			b.Errorf("error stopping the worker: error %v", err)
		}
	}
}

func Benchmark_Pipe_Worker_ExecEcho(b *testing.B) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(context.Background(), cmd)
	sw, err := workerImpl.From(w)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	go func() {
		err := w.Wait()
		if err != nil {
			b.Errorf("error waiting the worker: error %v", err)
		}
	}()
	defer func() {
		err := w.Stop(context.Background())
		if err != nil {
			b.Errorf("error stopping the worker: error %v", err)
		}
	}()

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEcho3(b *testing.B) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		err = w.Stop(ctx)
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := workerImpl.From(w)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEchoWithoutContext(b *testing.B) {
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		err = w.Stop(ctx)
		if err != nil {
			b.Errorf("error stopping the Process: error %v", err)
		}
	}()

	sw, err := workerImpl.From(w)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Test_Echo(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	syncWorker, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		assert.NoError(t, syncWorker.Wait())
	}()
	defer func() {
		err := syncWorker.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	res, err := syncWorker.Exec(payload.Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_BadPayload(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)

	syncWorker, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		assert.NoError(t, syncWorker.Wait())
	}()
	defer func() {
		err := syncWorker.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	res, err := syncWorker.Exec(payload.Payload{})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Contains(t, err.Error(), "payload can not be empty")
}

func Test_String(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	assert.Contains(t, w.String(), "php ../../tests/client.php echo pipes")
	assert.Contains(t, w.String(), "ready")
	assert.Contains(t, w.String(), "numExecs: 0")
}

func Test_Echo_Slow(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/slow-client.php", "echo", "pipes", "10", "10")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	syncWorker, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := syncWorker.Exec(payload.Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Broken(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "broken", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}

	data := ""
	mu := &sync.Mutex{}
	w.AddListener(func(event interface{}) {
		if wev, ok := event.(events.WorkerEvent); ok {
			mu.Lock()
			data = string(wev.Payload.([]byte))
			mu.Unlock()
		}
	})

	syncWorker, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := syncWorker.Exec(payload.Payload{Body: []byte("hello")})
	assert.NotNil(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	time.Sleep(time.Second * 3)
	mu.Lock()
	if strings.ContainsAny(data, "undefined_function()") == false {
		t.Fail()
	}
	mu.Unlock()
	assert.Error(t, w.Stop(ctx))
}

func Test_Error(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "error", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()

	defer func() {
		err := w.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	syncWorker, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := syncWorker.Exec(payload.Payload{Body: []byte("hello")})
	assert.NotNil(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	if errors.Is(errors.ErrSoftJob, err) == false {
		t.Fatal("error should be of type errors.ErrSoftJob")
	}
	assert.Contains(t, err.Error(), "hello")
}

func Test_NumExecs(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	syncWorker, err := workerImpl.From(w)
	if err != nil {
		t.Fatal(err)
	}

	_, err = syncWorker.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, int64(1), w.State().NumExecs())

	_, err = syncWorker.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, int64(2), w.State().NumExecs())

	_, err = syncWorker.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, int64(3), w.State().NumExecs())
}
