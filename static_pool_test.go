package roadrunner

import (
	"context"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var cfg = Config{
	NumWorkers:      int64(runtime.NumCPU()),
	AllocateTimeout: time.Second,
	DestroyTimeout:  time.Second,
}

func Test_NewPool(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)
}

func Test_StaticPool_Invalid(t *testing.T) {
	p, err := NewPool(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "tests/invalid.php") },
		NewPipeFactory(),
		cfg,
	)

	assert.Nil(t, p)
	assert.Error(t, err)
}

func Test_ConfigNoErrorInitDefaults(t *testing.T) {
	p, err := NewPool(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NotNil(t, p)
	assert.NoError(t, err)
}

func Test_StaticPool_Echo(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_StaticPool_Echo_NilContext(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(Payload{Body: []byte("hello"), Context: nil})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_StaticPool_Echo_Context(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "head", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(Payload{Body: []byte("hello"), Context: []byte("world")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Nil(t, res.Body)
	assert.NotNil(t, res.Context)

	assert.Equal(t, "world", string(res.Context))
}

func Test_StaticPool_JobError(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "error", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.IsType(t, ExecError{}, err)
	assert.Equal(t, "hello", err.Error())
}

// TODO temporary commented, figure out later
// func Test_StaticPool_Broken_Replace(t *testing.T) {
//	ctx := context.Background()
//	p, err := NewPool(
//		ctx,
//		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "broken", "pipes") },
//		NewPipeFactory(),
//		cfg,
//	)
//	assert.NoError(t, err)
//	assert.NotNil(t, p)
//
//	wg := &sync.WaitGroup{}
//	wg.Add(1)
//	var i int64
//	atomic.StoreInt64(&i, 10)
//
//	p.AddListener(func(event interface{}) {
//
//	})
//
//	go func() {
//		for {
//			select {
//			case ev := <-p.Events():
//				wev := ev.Payload.(WorkerEvent)
//				if _, ok := wev.Payload.([]byte); ok {
//					assert.Contains(t, string(wev.Payload.([]byte)), "undefined_function()")
//					wg.Done()
//					return
//				}
//			}
//		}
//	}()
//	res, err := p.ExecWithContext(ctx, Payload{Body: []byte("hello")})
//	assert.Error(t, err)
//	assert.Nil(t, res.Context)
//	assert.Nil(t, res.Body)
//	wg.Wait()
//
//	p.Destroy(ctx)
// }

//
func Test_StaticPool_Broken_FromOutside(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Nil(t, res.Context)

	assert.Equal(t, "hello", res.String())
	assert.Equal(t, runtime.NumCPU(), len(p.Workers()))

	// Consume pool events
	wg := sync.WaitGroup{}
	wg.Add(1)
	p.AddListener(func(event interface{}) {
		if pe, ok := event.(PoolEvent); ok {
			if pe.Event == EventWorkerConstruct {
				wg.Done()
			}
		}
	})

	// killing random worker and expecting pool to replace it
	err = p.Workers()[0].Kill(ctx)
	if err != nil {
		t.Errorf("error killing the process: error %v", err)
	}

	wg.Wait()

	list := p.Workers()
	for _, w := range list {
		assert.Equal(t, StateReady, w.State().Value())
	}
	wg.Wait()
}

func Test_StaticPool_AllocateTimeout(t *testing.T) {
	p, err := NewPool(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "delay", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Nanosecond * 1,
			DestroyTimeout:  time.Second * 2,
		},
	)
	assert.Error(t, err)
	assert.Nil(t, p)
}

func Test_StaticPool_Replace_Worker(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "pid", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			MaxJobs:         1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	var lastPID string
	lastPID = strconv.Itoa(int(p.Workers()[0].Pid()))

	res, _ := p.Exec(Payload{Body: []byte("hello")})
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Nil(t, res.Context)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

func Test_StaticPool_Debug_Worker(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "pid", "pipes") },
		NewPipeFactory(),
		Config{
			Debug:           true,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	assert.Len(t, p.Workers(), 0)

	var lastPID string
	res, _ := p.Exec(Payload{Body: []byte("hello")})
	assert.NotEqual(t, lastPID, string(res.Body))

	assert.Len(t, p.Workers(), 0)

	for i := 0; i < 10; i++ {
		assert.Len(t, p.Workers(), 0)
		res, err := p.Exec(Payload{Body: []byte("hello")})

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
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "stop", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	var lastPID string
	lastPID = strconv.Itoa(int(p.Workers()[0].Pid()))

	res, err := p.Exec(Payload{Body: []byte("hello")})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(Payload{Body: []byte("hello")})

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
	ctx := context.Background()
	p, err := NewPool(
		ctx,
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

	p.Destroy(ctx)
	_, err = p.Exec(Payload{Body: []byte("100")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Destroy_And_Close_While_Wait(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
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

	go func() {
		_, err := p.Exec(Payload{Body: []byte("100")})
		if err != nil {
			t.Errorf("error executing payload: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 10)

	p.Destroy(ctx)
	_, err = p.Exec(Payload{Body: []byte("100")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Handle_Dead(t *testing.T) {
	ctx := context.Background()
	p, err := NewPool(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "tests/slow-destroy.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      5,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	for _, w := range p.Workers() {
		w.State().Set(StateErrored)
	}

	_, err = p.Exec(Payload{Body: []byte("hello")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Slow_Destroy(t *testing.T) {
	p, err := NewPool(
		context.Background(),
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

	p.Destroy(context.Background())
}

func Benchmark_Pool_Echo(b *testing.B) {
	ctx := context.Background()
	p, err := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		cfg,
	)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

//
func Benchmark_Pool_Echo_Batched(b *testing.B) {
	ctx := context.Background()
	p, _ := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      int64(runtime.NumCPU()),
			AllocateTimeout: time.Second * 100,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy(ctx)

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := p.Exec(Payload{Body: []byte("hello")}); err != nil {
				b.Fail()
				log.Println(err)
			}
		}()
	}

	wg.Wait()
}

//
func Benchmark_Pool_Echo_Replaced(b *testing.B) {
	ctx := context.Background()
	p, _ := NewPool(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "tests/client.php", "echo", "pipes") },
		NewPipeFactory(),
		Config{
			NumWorkers:      1,
			MaxJobs:         1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	defer p.Destroy(ctx)
	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
			log.Println(err)
		}
	}
}
