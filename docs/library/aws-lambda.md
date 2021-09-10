# AWS Lambda
RoadRunner can run PHP as AWS Lambda function.

### Installation
Prior to the function deployment, you must compile or download PHP binary files to run your application. There are multiple projects available for such goal:

- https://github.com/araines/serverless-php
- https://github.com/stechstudio/php-lambda (includes pre-built versions of PHP binaries)

Place PHP binaries in a `bin/` folder of your project.

### PHP Worker
PHP worker does not require any specific configuration to run inside Lambda function. We can use default snippet with internal counter to demonstrate how workers are being reused:

```php
<?php
/**
 * @var Goridge\RelayInterface $relay
 */
use Spiral\Goridge;
use Spiral\RoadRunner;

ini_set('display_errors', 'stderr');
require __DIR__ . "/vendor/autoload.php";

$worker = new RoadRunner\Worker(new Goridge\StreamRelay(STDIN, STDOUT));
$psr7 = new RoadRunner\Http\PSR7Worker(
    $worker,
    new \Nyholm\Psr7\Factory\Psr17Factory(),
    new \Nyholm\Psr7\Factory\Psr17Factory(),
    new \Nyholm\Psr7\Factory\Psr17Factory()
);

while ($req = $psr7->waitRequest()) {
    try {
        $resp = new \Nyholm\Psr7\Response();
        $resp->getBody()->write(str_repeat("hello world", 1000));

        $psr7->respond($resp);
    } catch (\Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }

```

Name this file `handler.php` and put it into the root of your project. Make sure to run `composer require spiral/roadrunner`.

### Application
We can create a simple application to demonstrate how it works:
1. You need 3 files, main.go with the `Endure` container:

```golang
import (
	_ "embed"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/plugins/config"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
)

//go:embed .rr.yaml
var rrYaml []byte

func main() {
	_ = os.Setenv("PATH", os.Getenv("PATH")+":"+os.Getenv("LAMBDA_TASK_ROOT"))
	_ = os.Setenv("LD_LIBRARY_PATH", "./lib:/lib64:/usr/lib64")

	cont, err := endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel))
	if err != nil {
		log.Fatal(err)
	}

	cfg := &config.Viper{}
	cfg.CommonConfig = &config.General{GracefulTimeout: time.Second * 30}
	cfg.ReadInCfg = rrYaml
	cfg.Type = "yaml"

	// only 4 plugins needed here
	// 1. Server which should provide pool to us
	// 2. Our mini plugin, which expose this pool for us
	// 3. Logger
	// 4. Configurer
	err = cont.RegisterAll(
		cfg,
		&logger.ZapLogger{},
		&Plugin{},
		&server.Plugin{},
	)
	if err != nil {
		log.Fatal(err)
	}

	err = cont.Init()
	if err != nil {
		log.Fatal(err)
	}

	ch, err := cont.Serve()
	if err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case e := <-ch:
				log.Println(e.Error.Error())
			case <-sig:
				err = cont.Stop()
				if err != nil {
					log.Println(err)
				}
				return
			}
		}
	}()

	wg.Wait()
}
```

2. And `Plugin` for the RR:
```golang
package main

import (
	"context"
	"sync"
	"unsafe"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spiral/roadrunner/v2/pkg/payload"
	"github.com/spiral/roadrunner/v2/pkg/pool"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/server"
)

type Plugin struct {
	sync.Mutex
	log     logger.Logger
	srv     server.Server
	wrkPool pool.Pool
}

func (p *Plugin) Init(srv server.Server, log logger.Logger) error {
	var err error
	p.srv = srv
	p.log = log
	return err
}

func (p *Plugin) Serve() chan error {
	errCh := make(chan error, 1)
	p.Lock()
	defer p.Unlock()
	var err error

	p.wrkPool, err = p.srv.NewWorkerPool(context.Background(), pool.Config{
		Debug:           false,
		NumWorkers:      1,
		MaxJobs:         0,
		AllocateTimeout: 0,
		DestroyTimeout:  0,
		Supervisor: &pool.SupervisorConfig{
			WatchTick:       0,
			TTL:             0,
			IdleTTL:         0,
			ExecTTL:         0,
			MaxWorkerMemory: 0,
		},
	}, nil, nil)

	go func() {
		// register handler
		lambda.Start(p.handler())
	}()

	if err != nil {
		errCh <- err
	}
	return errCh
}

func (p *Plugin) Stop() error {
	p.Lock()
	defer p.Unlock()

	if p.wrkPool != nil {
		p.wrkPool.Destroy(context.Background())
	}
	return nil
}

func (p *Plugin) handler() func(pld string) (string, error) {
	return func(pld string) (string, error) {
		data := fastConvert(pld)
		// execute on worker pool
		if p.wrkPool == nil {
			// or any error
			return "", nil
		}
		exec, err := p.wrkPool.Exec(payload.Payload{
			Context: nil,
			Body:    data,
		})
		if err != nil {
			return "", err
		}
		return exec.String(), nil
	}
}

// reinterpret_cast conversion cast from string to []byte
// unsafe
func fastConvert(d string) []byte {
	return *(*[]byte)(unsafe.Pointer(&d))
}
```
3. Config file, which can be embedded into the binary with `embed` import:
```yaml
server:
  command: "php handler.php"
  user: ""
  group: ""
  env:
    - SOME_KEY: "SOME_VALUE"
    - SOME_KEY2: "SOME_VALUE2"
  relay: pipes
  relay_timeout: 60s
```
Here you can use full advantage of the RR2, you can include any plugin here and configure it with the embedded config (within reasonable limits).

To build and package your lambda function run:

```
$ GOOS=linux GOARCH=amd64 go build -o main main.go 
$ zip main.zip * -r
```

You can now upload and invoke your handler using simple string event.

## Notes
There are multiple notes you have to acknowledge.

- start with 1 worker per lambda function in order to control your memory usage.
- make sure to include env variables listed in the code to properly resolve the location of PHP binary and it's dependencies.
- avoid database connections without concurrency limit
- avoid database connections
