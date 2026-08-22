package container

import (
	"github.com/roadrunner-server/amqp/v6"
	appLogger "github.com/roadrunner-server/app-logger/v6"
	"github.com/roadrunner-server/beanstalk/v6"
	"github.com/roadrunner-server/boltdb/v6"
	"github.com/roadrunner-server/centrifuge/v6"
	"github.com/roadrunner-server/fileserver/v6"
	gps "github.com/roadrunner-server/google-pub-sub/v6"
	grpcPlugin "github.com/roadrunner-server/grpc/v6"
	"github.com/roadrunner-server/gzip/v6"
	"github.com/roadrunner-server/headers/v6"
	httpPlugin "github.com/roadrunner-server/http/v6"
	"github.com/roadrunner-server/informer/v6"
	"github.com/roadrunner-server/jobs/v6"
	"github.com/roadrunner-server/kafka/v6"
	"github.com/roadrunner-server/kv/v6"
	"github.com/roadrunner-server/lock/v6"
	"github.com/roadrunner-server/logger/v6"
	"github.com/roadrunner-server/memcached/v6"
	"github.com/roadrunner-server/memory/v6"
	"github.com/roadrunner-server/metrics/v6"
	"github.com/roadrunner-server/nats/v6"
	"github.com/roadrunner-server/nsq/v6"
	rrOtel "github.com/roadrunner-server/otel/v6"
	"github.com/roadrunner-server/prometheus/v6"
	proxyIP "github.com/roadrunner-server/proxy_ip_parser/v6"
	"github.com/roadrunner-server/redis/v6"
	"github.com/roadrunner-server/resetter/v6"
	rpcPlugin "github.com/roadrunner-server/rpc/v6"
	"github.com/roadrunner-server/send/v6"
	"github.com/roadrunner-server/server/v6"
	"github.com/roadrunner-server/service/v6"
	"github.com/roadrunner-server/sqs/v6"
	"github.com/roadrunner-server/static/v6"
	"github.com/roadrunner-server/status/v6"
	rrt "github.com/temporalio/roadrunner-temporal/v6"
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
		&nsq.Plugin{},
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
		&fileserver.Plugin{},
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
		// temporal plugin
		&rrt.Plugin{},
	}
}
