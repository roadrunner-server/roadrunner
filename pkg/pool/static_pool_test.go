package pool

import (
	"context"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spiral/errors"
	"github.com/spiral/roadrunner/v2/pkg/events"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/transport/pipe"
	"github.com/spiral/roadrunner/v2/pkg/worker"
	"github.com/spiral/roadrunner/v2/utils"
	"github.com/stretchr/testify/assert"
)

var cfg = Config{
	NumWorkers:      uint64(runtime.NumCPU()),
	AllocateTimeout: time.Second * 5,
	DestroyTimeout:  time.Second * 5,
}

func Test_NewPool(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)
}

func Test_StaticPool_Invalid(t *testing.T) {
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../tests/invalid.php") },
		pipe.NewPipeFactory(),
		cfg,
	)

	assert.Nil(t, p)
	assert.Error(t, err)
}

func Test_ConfigNoErrorInitDefaults(t *testing.T) {
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
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
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_StaticPool_Echo_NilContext(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(payload.Payload{Body: []byte("hello"), Context: nil})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
}

func Test_StaticPool_Echo_Context(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "head", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)

	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(payload.Payload{Body: []byte("hello"), Context: []byte("world")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Empty(t, res.Body)
	assert.NotNil(t, res.Context)

	assert.Equal(t, "world", string(res.Context))
}

func Test_StaticPool_JobError(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "error", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
	)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	res, err := p.Exec(payload.Payload{Body: []byte("hello")})

	assert.Error(t, err)
	assert.Nil(t, res.Body)
	assert.Nil(t, res.Context)

	if errors.Is(errors.SoftJob, err) == false {
		t.Fatal("error should be of type errors.Exec")
	}

	assert.Contains(t, err.Error(), "hello")
	p.Destroy(ctx)
}

func Test_StaticPool_Broken_Replace(t *testing.T) {
	ctx := context.Background()
	block := make(chan struct{}, 10)

	listener := func(event interface{}) {
		if wev, ok := event.(events.WorkerEvent); ok {
			if wev.Event == events.EventWorkerStderr {
				e := string(wev.Payload.([]byte))
				if strings.ContainsAny(e, "undefined_function()") {
					block <- struct{}{}
					return
				}
			}
		}
	}

	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "broken", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
		AddListeners(listener),
	)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	time.Sleep(time.Second)
	res, err := p.execWithTTL(ctx, payload.Payload{Body: []byte("hello")})
	assert.Error(t, err)
	assert.Nil(t, res.Context)
	assert.Nil(t, res.Body)

	<-block

	p.Destroy(ctx)
}

func Test_StaticPool_Broken_FromOutside(t *testing.T) {
	ctx := context.Background()
	// Consume pool events
	ev := make(chan struct{}, 1)
	listener := func(event interface{}) {
		if pe, ok := event.(events.PoolEvent); ok {
			if pe.Event == events.EventWorkerConstruct {
				ev <- struct{}{}
			}
		}
	}

	var cfg2 = Config{
		NumWorkers:      1,
		AllocateTimeout: time.Second * 5,
		DestroyTimeout:  time.Second * 5,
	}

	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		cfg2,
		AddListeners(listener),
	)
	assert.NoError(t, err)
	defer p.Destroy(ctx)

	assert.NotNil(t, p)

	res, err := p.Exec(payload.Payload{Body: []byte("hello")})

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotNil(t, res.Body)
	assert.Empty(t, res.Context)

	assert.Equal(t, "hello", res.String())
	assert.Equal(t, 1, len(p.Workers()))

	// first creation
	<-ev
	// killing random worker and expecting pool to replace it
	err = p.Workers()[0].Kill()
	if err != nil {
		t.Errorf("error killing the process: error %v", err)
	}

	// re-creation
	<-ev

	list := p.Workers()
	for _, w := range list {
		assert.Equal(t, worker.StateReady, w.State().Value())
	}
}

func Test_StaticPool_AllocateTimeout(t *testing.T) {
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "delay", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Nanosecond * 1,
			DestroyTimeout:  time.Second * 2,
		},
	)
	assert.Error(t, err)
	if !errors.Is(errors.WorkerAllocate, err) {
		t.Fatal("error should be of type WorkerAllocate")
	}
	assert.Nil(t, p)
}

func Test_StaticPool_Replace_Worker(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "pid", "pipes") },
		pipe.NewPipeFactory(),
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

	res, _ := p.Exec(payload.Payload{Body: []byte("hello")})
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(payload.Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Empty(t, res.Context)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

func Test_StaticPool_Debug_Worker(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "pid", "pipes") },
		pipe.NewPipeFactory(),
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
	res, _ := p.Exec(payload.Payload{Body: []byte("hello")})
	assert.NotEqual(t, lastPID, string(res.Body))

	assert.Len(t, p.Workers(), 0)

	for i := 0; i < 10; i++ {
		assert.Len(t, p.Workers(), 0)
		res, err := p.Exec(payload.Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Empty(t, res.Context)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

// identical to replace but controlled on worker side
func Test_StaticPool_Stop_Worker(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "stop", "pipes") },
		pipe.NewPipeFactory(),
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

	res, err := p.Exec(payload.Payload{Body: []byte("hello")})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, lastPID, string(res.Body))

	for i := 0; i < 10; i++ {
		res, err := p.Exec(payload.Payload{Body: []byte("hello")})

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.Body)
		assert.Empty(t, res.Context)

		assert.NotEqual(t, lastPID, string(res.Body))
		lastPID = string(res.Body)
	}
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Destroy_And_Close(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "delay", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NotNil(t, p)
	assert.NoError(t, err)

	p.Destroy(ctx)
	_, err = p.Exec(payload.Payload{Body: []byte("100")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Destroy_And_Close_While_Wait(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "delay", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.NotNil(t, p)
	assert.NoError(t, err)

	go func() {
		_, errP := p.Exec(payload.Payload{Body: []byte("100")})
		if errP != nil {
			t.Errorf("error executing payload: error %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 100)

	p.Destroy(ctx)
	_, err = p.Exec(payload.Payload{Body: []byte("100")})
	assert.Error(t, err)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Handle_Dead(t *testing.T) {
	ctx := context.Background()
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../tests/slow-destroy.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      5,
			AllocateTimeout: time.Second * 100,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	for i := range p.Workers() {
		p.Workers()[i].State().Set(worker.StateErrored)
	}

	_, err = p.Exec(payload.Payload{Body: []byte("hello")})
	assert.NoError(t, err)
	p.Destroy(ctx)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_Slow_Destroy(t *testing.T) {
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "../../tests/slow-destroy.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
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

func Test_StaticPool_NoFreeWorkers(t *testing.T) {
	ctx := context.Background()
	block := make(chan struct{}, 10)

	listener := func(event interface{}) {
		if ev, ok := event.(events.PoolEvent); ok {
			if ev.Event == events.EventNoFreeWorkers {
				block <- struct{}{}
			}
		}
	}

	p, err := Initialize(
		ctx,
		// sleep for the 3 seconds
		func() *exec.Cmd { return exec.Command("php", "../../tests/sleep.php", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			Debug:           false,
			NumWorkers:      1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
			Supervisor:      nil,
		},
		AddListeners(listener),
	)
	assert.NoError(t, err)
	assert.NotNil(t, p)

	go func() {
		_, _ = p.execWithTTL(ctx, payload.Payload{Body: []byte("hello")})
	}()

	time.Sleep(time.Second)
	res, err := p.execWithTTL(ctx, payload.Payload{Body: []byte("hello")})
	assert.Error(t, err)
	assert.Nil(t, res.Context)
	assert.Nil(t, res.Body)

	<-block

	p.Destroy(ctx)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_WrongCommand1(t *testing.T) {
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("phg", "../../tests/slow-destroy.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      5,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.Error(t, err)
	assert.Nil(t, p)
}

// identical to replace but controlled on worker side
func Test_Static_Pool_WrongCommand2(t *testing.T) {
	p, err := Initialize(
		context.Background(),
		func() *exec.Cmd { return exec.Command("php", "", "echo", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      5,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)

	assert.Error(t, err)
	assert.Nil(t, p)
}

func Benchmark_Pool_Echo(b *testing.B) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		cfg,
	)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
		}
	}
}

// Benchmark_Pool_Echo_Batched-32          366996          2873 ns/op        1233 B/op          24 allocs/op
func Benchmark_Pool_Echo_Batched(b *testing.B) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      uint64(runtime.NumCPU()),
			AllocateTimeout: time.Second * 100,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(b, err)
	defer p.Destroy(ctx)

	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := p.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
				b.Fail()
				log.Println(err)
			}
		}()
	}

	wg.Wait()
}

// Benchmark_Pool_Echo_Replaced-32    	     104/100	  10900218 ns/op	   52365 B/op	     125 allocs/op
func Benchmark_Pool_Echo_Replaced(b *testing.B) {
	ctx := context.Background()
	p, err := Initialize(
		ctx,
		func() *exec.Cmd { return exec.Command("php", "../../tests/client.php", "echo", "pipes") },
		pipe.NewPipeFactory(),
		Config{
			NumWorkers:      1,
			MaxJobs:         1,
			AllocateTimeout: time.Second,
			DestroyTimeout:  time.Second,
		},
	)
	assert.NoError(b, err)
	defer p.Destroy(ctx)
	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		if _, err := p.Exec(payload.Payload{Body: []byte("hello")}); err != nil {
			b.Fail()
			log.Println(err)
		}
	}
}

// BenchmarkToStringUnsafe-12    	566317729	         1.91 ns/op	       0 B/op	       0 allocs/op
// BenchmarkToStringUnsafe-32    	1000000000	         0.4434 ns/op	       0 B/op	       0 allocs/op
func BenchmarkToStringUnsafe(b *testing.B) {
	testPayload := []byte("falsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtoj")
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		res := utils.AsString(testPayload)
		_ = res
	}
}

// BenchmarkToStringSafe-32    	 8017846	       182.5 ns/op	     896 B/op	       1 allocs/op
// inline BenchmarkToStringSafe-12    	28926276	        46.6 ns/op	     128 B/op	       1 allocs/op
func BenchmarkToStringSafe(b *testing.B) {
	testPayload := []byte("falsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtojfalsflasjlifjwpoihejfoiwejow{}{}{}{}jelfjasjfhwaopiehjtopwhtgohrgouahsgkljasdlfjasl;fjals;jdflkndgouwhetopwqhjtoj")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		res := toStringNotFun(testPayload)
		_ = res
	}
}

func toStringNotFun(data []byte) string {
	return string(data)
}
