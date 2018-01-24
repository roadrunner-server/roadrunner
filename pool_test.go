package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"os/exec"
	"runtime"
	"sync"
	"testing"
	"time"
	"strconv"
)

var cfg = Config{
	NumWorkers:      uint64(runtime.NumCPU()),
	AllocateTimeout: time.Second,
	DestroyTimeout:  time.Second,
}

func Test_NewPool(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)
}

func Test_ConfigError(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.Nil(t, p)
	assert.Error(t, err)
}

func Test_Pool_Echo(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	res, err := p.Exec(&Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Head)

	assert.Equal(t, "hello", res.String())
}

func Test_Pool_AllocateTimeout(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "delay", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Millisecond * 50,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NotNil(t, p)
	assert.NoError(t, err)

	done := make(chan interface{})
	go func() {
		_, err := p.Exec(&Payload{Body: []byte("100")})
		assert.NoError(t, err)
		close(done)
	}()

	// to ensure that worker is already busy
	time.Sleep(time.Millisecond * 10)

	_, err = p.Exec(&Payload{Body: []byte("10")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "worker timeout")

	<-done
	p.Destroy()
}

//todo: termiante

func Test_Pool_Replace_Worker(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "pid", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			MaxExecutions:   1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	var lastPID string
	lastPID = strconv.Itoa(*p.Workers()[0].Pid)

	res, err := p.Exec(&Payload{Body: []byte("hello")})
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(&Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Nil(t, res.Head)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

func Benchmark_Pool_Allocate(b *testing.B) {
	p, _ := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	for n := 0; n < b.N; n++ {
		w, err := p.allocateWorker()
		if err != nil {
			b.Fail()
		}

		p.free <- w
	}
}

func Benchmark_Pool_Echo(b *testing.B) {
	p, _ := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(&Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

func Benchmark_Pool_Echo_Batched(b *testing.B) {
	p, _ := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := p.Exec(&Payload{Body: []byte("hello")}); err != nil {
				b.Fail()
			}
		}()
	}

	wg.Wait()
}

func Benchmark_Pool_Echo_Replaced(b *testing.B) {
	p, _ := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			MaxExecutions:   1,
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
