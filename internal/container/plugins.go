package container

import (
	"github.com/spiral/roadrunner-plugins/v2/amqp"
	"github.com/spiral/roadrunner-plugins/v2/beanstalk"
	"github.com/spiral/roadrunner-plugins/v2/boltdb"
	"github.com/spiral/roadrunner-plugins/v2/broadcast"
	"github.com/spiral/roadrunner-plugins/v2/fileserver"
	grpcPlugin "github.com/spiral/roadrunner-plugins/v2/grpc"
	httpPlugin "github.com/spiral/roadrunner-plugins/v2/http"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/gzip"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/headers"
	newrelic "github.com/spiral/roadrunner-plugins/v2/http/middleware/new_relic"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/prometheus"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/static"
	"github.com/spiral/roadrunner-plugins/v2/http/middleware/websockets"
	"github.com/spiral/roadrunner-plugins/v2/informer"
	"github.com/spiral/roadrunner-plugins/v2/jobs"
	"github.com/spiral/roadrunner-plugins/v2/kv"
	"github.com/spiral/roadrunner-plugins/v2/logger"
	"github.com/spiral/roadrunner-plugins/v2/memcached"
	"github.com/spiral/roadrunner-plugins/v2/memory"
	"github.com/spiral/roadrunner-plugins/v2/metrics"
	"github.com/spiral/roadrunner-plugins/v2/nats"
	"github.com/spiral/roadrunner-plugins/v2/redis"
	"github.com/spiral/roadrunner-plugins/v2/reload"
	"github.com/spiral/roadrunner-plugins/v2/resetter"
	rpcPlugin "github.com/spiral/roadrunner-plugins/v2/rpc"
	"github.com/spiral/roadrunner-plugins/v2/server"
	"github.com/spiral/roadrunner-plugins/v2/service"
	"github.com/spiral/roadrunner-plugins/v2/sqs"
	"github.com/spiral/roadrunner-plugins/v2/status"
	"github.com/spiral/roadrunner-plugins/v2/tcp"
	roadrunner_temporal "github.com/temporalio/roadrunner-temporal"
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
		&logger.ZapLogger{},
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
		&newrelic.Plugin{},
		&static.Plugin{},
		&headers.Plugin{},
		&status.Plugin{},
		&gzip.Plugin{},
		&prometheus.Plugin{},

		&fileserver.Plugin{},
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
		&roadrunner_temporal.Plugin{},
	}
}
