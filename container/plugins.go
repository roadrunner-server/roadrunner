package container

import (
	"github.com/roadrunner-server/amqp/v2"
	"github.com/roadrunner-server/beanstalk/v2"
	"github.com/roadrunner-server/boltdb/v2"
	"github.com/roadrunner-server/broadcast/v2"
	"github.com/roadrunner-server/cache/v2"
	"github.com/roadrunner-server/fileserver/v2"
	grpcPlugin "github.com/roadrunner-server/grpc/v2"
	"github.com/roadrunner-server/gzip/v2"
	"github.com/roadrunner-server/headers/v2"
	httpPlugin "github.com/roadrunner-server/http/v2"
	"github.com/roadrunner-server/informer/v2"
	"github.com/roadrunner-server/jobs/v2"
	"github.com/roadrunner-server/logger/v2"
	"github.com/roadrunner-server/memory/v2"
	"github.com/roadrunner-server/metrics/v2"
	"github.com/roadrunner-server/nats/v2"
	rrOtel "github.com/roadrunner-server/otel/v2"
	"github.com/roadrunner-server/prometheus/v2"
	proxyIP "github.com/roadrunner-server/proxy_ip_parser/v2"
	"github.com/roadrunner-server/redis/v2"
	"github.com/roadrunner-server/reload/v2"
	"github.com/roadrunner-server/resetter/v2"
	rpcPlugin "github.com/roadrunner-server/rpc/v2"
	"github.com/roadrunner-server/send/v2"
	"github.com/roadrunner-server/server/v2"
	"github.com/roadrunner-server/service/v2"
	"github.com/roadrunner-server/sqs/v2"
	"github.com/roadrunner-server/static/v2"
	"github.com/roadrunner-server/status/v2"
	"github.com/roadrunner-server/websockets/v2"

	"github.com/roadrunner-server/kv/v2"
	"github.com/roadrunner-server/memcached/v2"
	"github.com/roadrunner-server/tcp/v2"
	rrt "github.com/temporalio/roadrunner-temporal"
)

// Plugins returns active plugins for the endure container. Feel free to add or remove any plugins.
func Plugins() []interface{} { //nolint:funlen
	return []interface{}{
		// bundled
		// informer plugin (./rr workers, ./rr workers -i)
		&informer.Plugin{},
		// resetter plugin (./rr reset)
		&resetter.Plugin{},

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

		// ========= JOBS bundle
		&jobs.Plugin{},
		&amqp.Plugin{},
		&sqs.Plugin{},
		&nats.Plugin{},
		&beanstalk.Plugin{},
		// =========

		// http server plugin with middleware
		&httpPlugin.Plugin{},
		&static.Plugin{},
		&headers.Plugin{},
		&status.Plugin{},
		&gzip.Plugin{},
		&prometheus.Plugin{},
		&cache.Plugin{},
		&send.Plugin{},
		&proxyIP.Plugin{},
		&fileserver.Plugin{},
		&rrOtel.Plugin{},
		// ===================

		&grpcPlugin.Plugin{},
		// kv + ws + jobs plugin
		&memory.Plugin{},
		// KV + Jobs
		&boltdb.Plugin{},

		// broadcast via memory or redis
		// used in conjunction with Websockets, memory and redis plugins
		&broadcast.Plugin{},
		// ======== websockets broadcast bundle
		&websockets.Plugin{},
		&redis.Plugin{},
		// =========

		// ============== KV
		&kv.Plugin{},
		&memcached.Plugin{},
		// ==============

		// raw TCP connections handling
		&tcp.Plugin{},

		// temporal plugins
		&rrt.Plugin{},
	}
}
