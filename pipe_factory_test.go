package roadrunner

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Pipe_Start(t *testing.T) {
	ctx := context.Background()
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		ctx := context.Background()
		assert.NoError(t, w.Wait(ctx))
	}()

	assert.NoError(t, w.Stop(ctx))
}

func Test_Pipe_StartError(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
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
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
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
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
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
	cmd := exec.Command("php", "tests/failboot.php")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)

	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failboot")
}

func Test_Pipe_Invalid(t *testing.T) {
	cmd := exec.Command("php", "tests/invalid.php")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Echo(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = w.Stop(ctx)
		if err != nil {
			t.Errorf("error stopping the WorkerProcess: error %v", err)
		}
	}()

	sw, err := NewSyncWorker(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.ExecWithContext(ctx, Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_Pipe_Broken(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "broken", "pipes")
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

	sw, err := NewSyncWorker(w)
	if err != nil {
		t.Fatal(err)
	}

	res, err := sw.ExecWithContext(ctx, Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)
}

func Benchmark_Pipe_SpawnWorker_Stop(b *testing.B) {
	f := NewPipeFactory()
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
		w, _ := f.SpawnWorkerWithContext(context.Background(), cmd)
		go func() {
			if w.Wait(context.Background()) != nil {
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
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorkerWithContext(context.Background(), cmd)
	sw, err := NewSyncWorker(w)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	go func() {
		err := w.Wait(context.Background())
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
		if _, err := sw.ExecWithContext(context.Background(), Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEcho3(b *testing.B) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		err = w.Stop(ctx)
		if err != nil {
			b.Errorf("error stopping the WorkerProcess: error %v", err)
		}
	}()

	sw, err := NewSyncWorker(w)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		if _, err := sw.ExecWithContext(ctx, Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pipe_Worker_ExecEchoWithoutContext(b *testing.B) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	ctx := context.Background()
	w, err := NewPipeFactory().SpawnWorkerWithContext(ctx, cmd)
	if err != nil {
		b.Fatal(err)
	}

	defer func() {
		err = w.Stop(ctx)
		if err != nil {
			b.Errorf("error stopping the WorkerProcess: error %v", err)
		}
	}()

	sw, err := NewSyncWorker(w)
	if err != nil {
		b.Fatal(err)
	}

	for n := 0; n < b.N; n++ {
		if _, err := sw.Exec(Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}
