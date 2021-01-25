package main

import (
	"log"

	endure "github.com/spiral/endure/pkg/container"
	"github.com/spiral/roadrunner/v2/cmd/cli"
	httpPlugin "github.com/spiral/roadrunner/v2/plugins/http"
	"github.com/spiral/roadrunner/v2/plugins/informer"
	"github.com/spiral/roadrunner/v2/plugins/temporal/activity"
	temporalClient "github.com/spiral/roadrunner/v2/plugins/temporal/client"
	"github.com/spiral/roadrunner/v2/plugins/temporal/workflow"

	"github.com/spiral/roadrunner/v2/plugins/kv/boltdb"
	"github.com/spiral/roadrunner/v2/plugins/kv/memcached"
	"github.com/spiral/roadrunner/v2/plugins/kv/memory"
	"github.com/spiral/roadrunner/v2/plugins/logger"
	"github.com/spiral/roadrunner/v2/plugins/metrics"
	"github.com/spiral/roadrunner/v2/plugins/redis"
	"github.com/spiral/roadrunner/v2/plugins/reload"
	"github.com/spiral/roadrunner/v2/plugins/resetter"
	"github.com/spiral/roadrunner/v2/plugins/rpc"
	"github.com/spiral/roadrunner/v2/plugins/server"
)

func main() {
	var err error
	cli.Container, err = endure.NewContainer(nil, endure.SetLogLevel(endure.ErrorLevel), endure.RetryOnFail(false))
	if err != nil {
		log.Fatal(err)
	}

	err = cli.Container.RegisterAll(
		// logger plugin
		&logger.ZapLogger{},
		// metrics plugin
		&metrics.Plugin{},
		// redis plugin (internal)
		&redis.Plugin{},
		// http server plugin
		&httpPlugin.Plugin{},
		// reload plugin
		&reload.Plugin{},
		// informer plugin (./rr workers, ./rr workers -i)
		&informer.Plugin{},
		// resetter plugin (./rr reset)
		&resetter.Plugin{},
		// rpc plugin (workers, reset)
		&rpc.Plugin{},
		// server plugin (NewWorker, NewWorkerPool)
		&server.Plugin{},
		// memcached kv plugin
		&memcached.Plugin{},
		// in-memory kv plugin
		&memory.Plugin{},
		// boltdb driver
		&boltdb.Plugin{},

		// temporal plugins
		&temporalClient.Plugin{},
		&activity.Plugin{},
		&workflow.Plugin{},
	)
	if err != nil {
		log.Fatal(err)
	}

	cli.Execute()
}
