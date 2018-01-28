package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

func Test_Pipe_Start(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.NoError(t, err)
	assert.NotNil(t, w)

	go func() {
		assert.NoError(t, w.Wait())
	}()

	w.Stop()
}

func Test_Pipe_StartError(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	cmd.Start()

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	cmd.StdinPipe()

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_PipeError2(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
	cmd.StdoutPipe()

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Failboot(t *testing.T) {
	cmd := exec.Command("php", "tests/failboot.php")
	w, err := NewPipeFactory().SpawnWorker(cmd)

	assert.Nil(t, w)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failboot")
}

func Test_Pipe_Invalid(t *testing.T) {
	cmd := exec.Command("php", "tests/invalid.php")

	w, err := NewPipeFactory().SpawnWorker(cmd)
	assert.Error(t, err)
	assert.Nil(t, w)
}

func Test_Pipe_Echo(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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

func Test_Pipe_Broken(t *testing.T) {
	cmd := exec.Command("php", "tests/client.php", "broken", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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

func Benchmark_Pipe_SpawnWorker_Stop(b *testing.B) {
	f := NewPipeFactory()
	for n := 0; n < b.N; n++ {
		cmd := exec.Command("php", "tests/client.php", "echo", "pipes")
		w, _ := f.SpawnWorker(cmd)
		go func() {
			if w.Wait() != nil {
				b.Fail()
			}
		}()

		w.Stop()
	}
}

func Benchmark_Pipe_Worker_ExecEcho(b *testing.B) {
	cmd := exec.Command("php", "tests/client.php", "echo", "pipes")

	w, _ := NewPipeFactory().SpawnWorker(cmd)
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
