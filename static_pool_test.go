package roadrunner

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

var cfg = Config{
	NumWorkers:      int64(runtime.NumCPU()),
	AllocateTimeout: time.Second,
	DestroyTimeout:  time.Second,
}

func Test_NewPool(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.Equal(t, cfg, p.Config())

	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)
}

func Test_StaticPool_Invalid(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/invalid.php") },
		NewPipeFactory(),
		cfg,
	)

	assert.Nil(t, p)
	assert.Error(t, err)
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

func Test_StaticPool_Echo(t *testing.T) {
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
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_StaticPool_Echo_NilContext(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	res, err := p.Exec(&Payload{Body: []byte("hello"), Context: nil})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_StaticPool_Echo_Context(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "head", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	res, err := p.Exec(&Payload{Body: []byte("hello"), Context: []byte("world")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Nil(t, res.Body)
	assert.NotNil(t, res.Context)

	assert.Equal(t, "world", string(res.Context))
}

func Test_StaticPool_JobError(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "error", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	res, err := p.Exec(&Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res)

	assert.IsType(t, JobError{}, err)
	assert.Equal(t, "hello", err.Error())
}

func Test_StaticPool_Broken_Replace(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "broken", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	p.Listen(func(e int, ctx interface{}) {
		if err, ok := ctx.(error); ok {
			assert.Contains(t, err.Error(), "undefined_function()")
		}
	})

	res, err := p.Exec(&Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res)
}

func Test_StaticPool_Broken_FromOutside(t *testing.T) {
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
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
	assert.Equal(t, runtime.NumCPU(), len(p.Workers()))

	destructed := make(chan interface{})
	p.Listen(func(e int, ctx interface{}) {
		if e == EventWorkerConstruct {
			destructed <- nil
		}
	})

	// killing random worker and expecting pool to replace it
	p.muw.Lock()
	p.workers[0].cmd.Process.Kill()
	p.muw.Unlock()
	<-destructed

	for _, w := range p.Workers() {
		assert.Equal(t, StateReady, w.state.Value())
	}
}

func Test_StaticPool_AllocateTimeout(t *testing.T) {
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

func Test_StaticPool_Replace_Worker(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "pid", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			MaxJobs:         1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	var lastPID string
	lastPID = strconv.Itoa(*p.Workers()[0].Pid)

	res, _ := p.Exec(&Payload{Body: []byte("hello")})
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(&Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Nil(t, res.Context)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

// identical to replace but controlled on worker side
func Test_StaticPool_Stop_Worker(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "stop", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	var lastPID string
	lastPID = strconv.Itoa(*p.Workers()[0].Pid)

	res, _ := p.Exec(&Payload{Body: []byte("hello")})
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(&Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Nil(t, res.Context)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Destroy_And_Close(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "delay", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NotNil(t, p)
	assert.NoError(t, err)

	p.Destroy()
	_, err = p.Exec(&Payload{Body: []byte("100")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Destroy_And_Close_While_Wait(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "delay", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NotNil(t, p)
	assert.NoError(t, err)

	go p.Exec(&Payload{Body: []byte("100")})
	time.Sleep(time.Millisecond * 10)

	p.Destroy()
	_, err = p.Exec(&Payload{Body: []byte("100")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Handle_Dead(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      5,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	assert.NotNil(t, p)
	assert.NoError(t, err)

	for _, w := range p.workers {
		w.state.value = StateErrored
	}

	_, err = p.Exec(&Payload{Body: []byte("hello")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Slow_Destroy(t *testing.T) {
	p, err := NewPool(
		func() *exec.Cmd { return exec.Command("php", "tests/slow-destroy.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      5,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NoError(t, err)
	assert.NotNil(t, p)

	p.Destroy()
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
			log.Println(err)
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
		Config{
			NumWorkers:      int64(runtime.NumCPU()),
			AllocateTimeout: time.Second * 100,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := p.Exec(&Payload{Body: []byte("hello")}); err != nil {
				b.Fail()
				log.Println(err)
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
			MaxJobs:         1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy()

	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(&Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
			log.Println(err)
		}
	}
}
