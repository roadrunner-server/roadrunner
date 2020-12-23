package main

import (
	"log"

	"github.com/spiral/endure"
	"github.com/spiral/roadrunner-plugins/http"
	"github.com/spiral/roadrunner-plugins/informer"
	"github.com/spiral/roadrunner-plugins/logger"
	"github.com/spiral/roadrunner-plugins/metrics"
	"github.com/spiral/roadrunner-plugins/redis"
	"github.com/spiral/roadrunner-plugins/reload"
	"github.com/spiral/roadrunner-plugins/resetter"
	"github.com/spiral/roadrunner-plugins/rpc"
	"github.com/spiral/roadrunner-plugins/server"
	"github.com/spiral/roadrunner/v2/cmd/cli"
)

func main() {
	var err error
	cli.Container, err = endure.NewContainer(nil, endure.SetLogLevel(endure.DebugLevel), endure.RetryOnFail(false))
	if err != nil {
		return
	}

	err = cli.Container.RegisterAll(
		// logger plugin
		&logger.ZapLogger{},
		// metrics plugin
		&metrics.Plugin{},
		// redis plugin (internal)
		&redis.Plugin{},
		// http server plugin
		&http.Plugin{},
		// reload plugin
		&reload.Plugin{},
		// informer plugin (./rr workers)
		&informer.Plugin{},
		// resetter plugin (./rr reset)
		&resetter.Plugin{},
		// rpc plugin (workers, reset)
		&rpc.Plugin{},
		// server plugin (NewWorker, NewWorkerPool)
		&server.Plugin{},
	)
	if err != nil {
		log.Fatal(err)
	}

	cli.Execute()
}
