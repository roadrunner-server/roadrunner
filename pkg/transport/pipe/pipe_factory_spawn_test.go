package pipe

import (
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/stretchr/testify/assert"
)

func Test_GetState2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
		assert.Equal(t, worker.StateStopped, w.State().Value())
	}()

	assert.NoError(t, err)
	assert.NotNil(t, w)

	assert.Equal(t, worker.StateReady, w.State().Value())
	assert.NoError(t, w.Stop())
}

func Test_Kill2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
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

func Test_Pipe_Start2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	assert.NoError(t, w.Stop())
}

func Test_Pipe_StartError2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
	err := cmd.Start()
	if err != nil {
		t.Errorf("error running the command: error %v", err)
	}

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError3(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
	_, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("error creating the STDIN pipe: error %v", err)
	}

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError4(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
	_, err := cmd.StdinPipe()
	if err != nil {
		t.Errorf("error creating the STDIN pipe: error %v", err)
	}

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Failboot2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/failboot.php")
	finish := make(chan struct{}, 10)
	listener := func(event interface{}) {
		if ev, ok := event.(events.WorkerEvent); ok {
			if ev.Event == events.EventWorkerStderr {
				if strings.Contains(string(ev.Payload.([]byte)), "failboot") {
					finish <- struct{}{}
				}
			}
		}
	}
	w, err := NewPipeFactory().SpawnWorker(cmd, listener)

	assert.Nil(t, w)
	assert.Error(t, err)
	<-finish
}

func Test_Pipe_Invalid2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/invalid.php")
	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Echo2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
	w, err := NewPipeFactory().SpawnWorker(cmd)
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

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Pipe_Broken2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "broken", "pipes")
	w, err := NewPipeFactory().SpawnWorker(cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		time.Sleep(time.Second)
		err = w.Stop()
		assert.Error(t, err)
	}()

	sw := worker.From(w)

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)
}

func Benchmark_Pipe_SpawnWorker_Stop2(b *testing.B) {
	f := NewPipeFactory()
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
		w, _ := f.SpawnWorker(cmd)
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

func Benchmark_Pipe_Worker_ExecEcho2(b *testing.B) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEcho4(b *testing.B) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
	w, err := NewPipeFactory().SpawnWorker(cmd)
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
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEchoWithoutContext2(b *testing.B) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")
	w, err := NewPipeFactory().SpawnWorker(cmd)
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
		if _, err := sw.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Test_Echo2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	if err != nil {
		t.Fatal(err)
	}

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

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_BadPayload2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)

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

	res, err := sw.Exec(payload.Payload{})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Contains(t, err.Error(), "payload can not be empty")
}

func Test_String2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
	go func() {
		assert.NoError(t, w.Wait())
	}()
	defer func() {
		err := w.Stop()
		if err != nil {
			t.Errorf("error stopping the Process: error %v", err)
		}
	}()

	assert.Contains(t, w.String(), "php ../../../tests/client.php echo pipes")
	assert.Contains(t, w.String(), "ready")
	assert.Contains(t, w.String(), "numExecs: 0")
}

func Test_Echo_Slow2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/slow-client.php", "echo", "pipes", "10", "10")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})

	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Broken2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "broken", "pipes")
	data := ""
	mu := &sync.Mutex{}
	listener := func(event interface{}) {
		if wev, ok := event.(events.WorkerEvent); ok {
			mu.Lock()
			data = string(wev.Payload.([]byte))
			mu.Unlock()
		}
	}

	w, err := NewPipeFactory().SpawnWorker(cmd, listener)
	if err != nil {
		t.Fatal(err)
	}

	sw := worker.From(w)

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})
	assert.NotNil(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	time.Sleep(time.Second * 3)
	mu.Lock()
	if strings.ContainsAny(data, "undefined_function()") == false {
		t.Fail()
	}
	mu.Unlock()
	assert.Error(t, w.Stop())
}

func Test_Error2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "error", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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

	res, err := sw.Exec(payload.Payload{Body: []byte("hello")})
	assert.NotNil(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	if errors.Is(errors.SoftJob, err) == false {
		t.Fatal("error should be of type errors.ErrSoftJob")
	}
	assert.Contains(t, err.Error(), "hello")
}

func Test_NumExecs2(t *testing.T) {
	cmd := exec.Command("php", "../../../tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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

	_, err := sw.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, uint64(1), w.State().NumExecs())

	_, err = sw.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, uint64(2), w.State().NumExecs())

	_, err = sw.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Errorf("fail to execute payload: error %v", err)
	}
	assert.Equal(t, uint64(3), w.State().NumExecs())
}
