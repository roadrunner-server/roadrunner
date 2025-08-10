package container

import (
	"github.com/roadrunner-server/amqp/v5"
	appLogger "github.com/roadrunner-server/app-logger/v5"
	"github.com/roadrunner-server/beanstalk/v5"
	"github.com/roadrunner-server/boltdb/v5"
	"github.com/roadrunner-server/centrifuge/v5"
	gps "github.com/roadrunner-server/google-pub-sub/v5"
	grpcPlugin "github.com/roadrunner-server/grpc/v5"
	"github.com/roadrunner-server/gzip/v5"
	"github.com/roadrunner-server/headers/v5"
	httpPlugin "github.com/roadrunner-server/http/v5"
	"github.com/roadrunner-server/informer/v5"
	"github.com/roadrunner-server/jobs/v5"
	"github.com/roadrunner-server/kafka/v5"
	"github.com/roadrunner-server/kv/v5"
	"github.com/roadrunner-server/lock/v5"
	"github.com/roadrunner-server/logger/v5"
	"github.com/roadrunner-server/memcached/v5"
	"github.com/roadrunner-server/memory/v5"
	"github.com/roadrunner-server/metrics/v5"
	"github.com/roadrunner-server/nats/v5"
	rrOtel "github.com/roadrunner-server/otel/v5"
	"github.com/roadrunner-server/prometheus/v5"
	proxyIP "github.com/roadrunner-server/proxy_ip_parser/v5"
	"github.com/roadrunner-server/redis/v5"
	"github.com/roadrunner-server/resetter/v5"
	rpcPlugin "github.com/roadrunner-server/rpc/v5"
	"github.com/roadrunner-server/send/v5"
	"github.com/roadrunner-server/server/v5"
	"github.com/roadrunner-server/service/v5"
	"github.com/roadrunner-server/sqs/v5"
	"github.com/roadrunner-server/static/v5"
	"github.com/roadrunner-server/status/v5"
	"github.com/roadrunner-server/tcp/v5"
	rrt "github.com/temporalio/roadrunner-temporal/v5"
)

// Plugins return active plugins for the endured container. Feel free to add or remove any plugins.
func Plugins() []any { //nolint:funlen
	return []any{
		// bundled
		// informer plugin (./rr workers, ./rr workers -i)
		&informer.Plugin{},
		// resetter plugin (./rr reset)
		&resetter.Plugin{},
		// mutexes(locks)
		&lock.Plugin{},
		// logger plugin
		&logger.Plugin{},
		// psr-3 logger extension
		&appLogger.Plugin{},
		// metrics plugin
		&metrics.Plugin{},
		// rpc plugin (workers, reset)
		&rpcPlugin.Plugin{},
		// server plugin (NewWorker, NewWorkerPool)
		&server.Plugin{},
		// service plugin
		&service.Plugin{},
		// centrifuge
		&centrifuge.Plugin{},
		//
		// ========= JOBS bundle
		&jobs.Plugin{},
		&amqp.Plugin{},
		&sqs.Plugin{},
		&nats.Plugin{},
		&kafka.Plugin{},
		&beanstalk.Plugin{},
		&gps.Plugin{},
		// =========
		//
		// http server plugin with middleware
		&httpPlugin.Plugin{},
		&static.Plugin{},
		&headers.Plugin{},
		&status.Plugin{},
		&gzip.Plugin{},
		&prometheus.Plugin{},
		&send.Plugin{},
		&proxyIP.Plugin{},
		&rrOtel.Plugin{},
		// ===================
		// gRPC
		&grpcPlugin.Plugin{},
		// ===================
		//  KV + Jobs
		&memory.Plugin{},
		//  KV + Jobs
		&boltdb.Plugin{},
		//  ============== KV
		&kv.Plugin{},
		&memcached.Plugin{},
		&redis.Plugin{},
		//  ==============
		// raw TCP connections handling
		&tcp.Plugin{},
		// temporal plugin
		&rrt.Plugin{},
	}
}
