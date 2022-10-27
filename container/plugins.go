package container

import (
	cache "github.com/darkweak/souin/plugins/roadrunner"
	"github.com/roadrunner-server/amqp/v3"
	"github.com/roadrunner-server/beanstalk/v3"
	"github.com/roadrunner-server/boltdb/v3"
	"github.com/roadrunner-server/centrifuge/v3"
	"github.com/roadrunner-server/fileserver/v3"
	grpcPlugin "github.com/roadrunner-server/grpc/v3"
	"github.com/roadrunner-server/gzip/v3"
	"github.com/roadrunner-server/headers/v3"
	httpPlugin "github.com/roadrunner-server/http/v3"
	"github.com/roadrunner-server/informer/v3"
	"github.com/roadrunner-server/jobs/v3"
	"github.com/roadrunner-server/kafka/v3"
	"github.com/roadrunner-server/kv/v3"
	"github.com/roadrunner-server/logger/v3"
	"github.com/roadrunner-server/memcached/v3"
	"github.com/roadrunner-server/memory/v3"
	"github.com/roadrunner-server/metrics/v3"
	"github.com/roadrunner-server/nats/v3"
	rrOtel "github.com/roadrunner-server/otel/v3"
	"github.com/roadrunner-server/prometheus/v3"
	proxyIP "github.com/roadrunner-server/proxy_ip_parser/v3"
	"github.com/roadrunner-server/redis/v3"
	"github.com/roadrunner-server/reload/v3"
	"github.com/roadrunner-server/resetter/v3"
	rpcPlugin "github.com/roadrunner-server/rpc/v3"
	"github.com/roadrunner-server/send/v3"
	"github.com/roadrunner-server/server/v3"
	"github.com/roadrunner-server/service/v3"
	"github.com/roadrunner-server/sqs/v3"
	"github.com/roadrunner-server/static/v3"
	"github.com/roadrunner-server/status/v3"
	"github.com/roadrunner-server/tcp/v3"
	rrt "github.com/temporalio/roadrunner-temporal/v2"
)

// Plugins returns active plugins for the endure container. Feel free to add or remove any plugins.
func Plugins() []any { //nolint:funlen
	return []any{
		// bundled
		// informer plugin (./rr workers, ./rr workers -i)
		&informer.Plugin{},
		// resetter plugin (./rr reset)
		&resetter.Plugin{},
		//
		// logger plugin
		&logger.Plugin{},
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
		// new in 2.11
		&kafka.Plugin{},
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
		// third-party--
		&cache.Plugin{},
	}
}
