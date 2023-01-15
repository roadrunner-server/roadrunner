package container

import (
	"github.com/roadrunner-server/amqp/v4"
	appLogger "github.com/roadrunner-server/app-logger/v4"
	"github.com/roadrunner-server/beanstalk/v4"
	"github.com/roadrunner-server/boltdb/v4"
	"github.com/roadrunner-server/centrifuge/v4"
	"github.com/roadrunner-server/fileserver/v4"
	grpcPlugin "github.com/roadrunner-server/grpc/v4"
	"github.com/roadrunner-server/gzip/v4"
	"github.com/roadrunner-server/headers/v4"
	httpPlugin "github.com/roadrunner-server/http/v4"
	"github.com/roadrunner-server/informer/v4"
	"github.com/roadrunner-server/jobs/v4"
	"github.com/roadrunner-server/kv/v4"
	"github.com/roadrunner-server/logger/v4"
	"github.com/roadrunner-server/memcached/v4"
	"github.com/roadrunner-server/memory/v4"
	"github.com/roadrunner-server/metrics/v4"
	"github.com/roadrunner-server/nats/v4"
	rrOtel "github.com/roadrunner-server/otel/v4"
	"github.com/roadrunner-server/prometheus/v4"
	proxyIP "github.com/roadrunner-server/proxy_ip_parser/v4"
	"github.com/roadrunner-server/redis/v4"
	"github.com/roadrunner-server/reload/v4"
	"github.com/roadrunner-server/resetter/v4"
	rpcPlugin "github.com/roadrunner-server/rpc/v4"
	"github.com/roadrunner-server/send/v4"
	"github.com/roadrunner-server/server/v4"
	"github.com/roadrunner-server/service/v4"
	"github.com/roadrunner-server/sqs/v4"
	"github.com/roadrunner-server/static/v4"
	"github.com/roadrunner-server/status/v4"
	"github.com/roadrunner-server/tcp/v4"
	rrt "github.com/temporalio/roadrunner-temporal/v4"
)

// Plugins returns active plugins for the endure container. Feel free to add or remove any plugins.
func Plugins() []any { //nolint:funlen
	return []any{
		// bundled
		// informer plugin (./rr workers, ./rr workers -i)
		&informer.Plugin{},
		// resetter plugin (./rr reset)
		&resetter.Plugin{},

		// logger plugin
		&logger.Plugin{},
		// psr-3 logger extension
		&appLogger.Plugin{},
		// metrics plugin
		&metrics.Plugin{},
		// reload plugin
		&reload.Plugin{},
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
		&beanstalk.Plugin{},
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
		// raw TCP connections handling
		&tcp.Plugin{},
		// temporal plugin
		&rrt.Plugin{},
	}
}
