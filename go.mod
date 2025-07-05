module github.com/roadrunner-server/roadrunner/v2025

go 1.24.1

toolchain go1.24.4

require (
	github.com/buger/goterm v1.0.4
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.18.0
	github.com/joho/godotenv v1.5.1
	github.com/olekukonko/tablewriter v1.0.8
	github.com/roadrunner-server/amqp/v5 v5.2.2
	github.com/roadrunner-server/api/v4 v4.20.0
	github.com/roadrunner-server/app-logger/v5 v5.1.8
	github.com/roadrunner-server/beanstalk/v5 v5.1.8
	github.com/roadrunner-server/boltdb/v5 v5.1.8
	github.com/roadrunner-server/centrifuge/v5 v5.1.8
	github.com/roadrunner-server/config/v5 v5.1.8
	github.com/roadrunner-server/endure/v2 v2.6.2
	github.com/roadrunner-server/errors v1.4.1
	github.com/roadrunner-server/fileserver/v5 v5.1.8
	github.com/roadrunner-server/google-pub-sub/v5 v5.1.8
	github.com/roadrunner-server/goridge/v3 v3.8.3
	github.com/roadrunner-server/grpc/v5 v5.2.2
	github.com/roadrunner-server/gzip/v5 v5.1.8
	github.com/roadrunner-server/headers/v5 v5.1.8
	github.com/roadrunner-server/http/v5 v5.2.7
	github.com/roadrunner-server/informer/v5 v5.1.8
	github.com/roadrunner-server/jobs/v5 v5.1.8
	github.com/roadrunner-server/kafka/v5 v5.2.4
	github.com/roadrunner-server/kv/v5 v5.2.8
	github.com/roadrunner-server/lock/v5 v5.1.8
	github.com/roadrunner-server/logger/v5 v5.1.8
	github.com/roadrunner-server/memcached/v5 v5.1.8
	github.com/roadrunner-server/memory/v5 v5.2.8
	github.com/roadrunner-server/metrics/v5 v5.1.8
	github.com/roadrunner-server/nats/v5 v5.1.8
	github.com/roadrunner-server/otel/v5 v5.2.0
	github.com/roadrunner-server/pool v1.1.3
	github.com/roadrunner-server/prometheus/v5 v5.1.7
	github.com/roadrunner-server/proxy_ip_parser/v5 v5.1.8
	github.com/roadrunner-server/redis/v5 v5.1.9
	github.com/roadrunner-server/resetter/v5 v5.1.8
	github.com/roadrunner-server/rpc/v5 v5.1.8
	github.com/roadrunner-server/send/v5 v5.1.5
	github.com/roadrunner-server/server/v5 v5.2.9
	github.com/roadrunner-server/service/v5 v5.1.8
	github.com/roadrunner-server/sqs/v5 v5.1.8
	github.com/roadrunner-server/static/v5 v5.1.6
	github.com/roadrunner-server/status/v5 v5.1.8
	github.com/roadrunner-server/tcp/v5 v5.1.8
	github.com/spf13/cobra v1.9.1
	github.com/spf13/viper v1.20.1
	github.com/stretchr/testify v1.10.0
	github.com/temporalio/roadrunner-temporal/v5 v5.7.0
	go.uber.org/automaxprocs v1.6.0
)

exclude (
	github.com/spf13/viper v1.18.0
	github.com/spf13/viper v1.18.1
	go.temporal.io/api v1.26.1
)

require (
	cloud.google.com/go v0.121.3 // indirect
	cloud.google.com/go/auth v0.16.2 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.7.0 // indirect
	cloud.google.com/go/iam v1.5.2 // indirect
	cloud.google.com/go/pubsub v1.49.0 // indirect
	github.com/andybalholm/brotli v1.2.0 // indirect
	github.com/aws/aws-sdk-go v1.55.7 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.5 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.17 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.70 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.32 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.36 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.38.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.0 // indirect
	github.com/aws/smithy-go v1.22.4 // indirect
	github.com/beanstalkd/go-beanstalk v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20250403215159-8d39553ac7cf // indirect
	github.com/cactus/go-statsd-client/v5 v5.1.0 // indirect
	github.com/caddyserver/certmagic v0.23.0 // indirect
	github.com/caddyserver/zerossl v0.1.3 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emicklei/proto v1.14.2 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.3.0 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/gofiber/fiber/v2 v2.52.8 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.6 // indirect
	github.com/googleapis/gax-go/v2 v2.14.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.11 // indirect
	github.com/libdns/libdns v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/mholt/acmez v1.2.0 // indirect
	github.com/mholt/acmez/v3 v3.1.2 // indirect
	github.com/miekg/dns v1.1.66 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.43.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nexus-rpc/sdk-go v0.4.0 // indirect
	github.com/olekukonko/errors v1.1.0 // indirect
	github.com/olekukonko/ll v0.0.9 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.22 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.22.0 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/quic-go/quic-go v0.53.0 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.11.0 // indirect
	github.com/redis/go-redis/extra/redisotel/v9 v9.11.0 // indirect
	github.com/redis/go-redis/extra/redisprometheus/v9 v9.11.0 // indirect
	github.com/redis/go-redis/v9 v9.11.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/roadrunner-server/context v1.1.0 // indirect
	github.com/roadrunner-server/events v1.0.1 // indirect
	github.com/roadrunner-server/priority_queue v1.0.5 // indirect
	github.com/roadrunner-server/tcplisten v1.5.2 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sagikazarmark/locafero v0.9.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.14.0 // indirect
	github.com/spf13/cast v1.9.2 // indirect
	github.com/spf13/pflag v1.0.6 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/twmb/franz-go v1.19.5 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.11.2 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/uber-go/tally/v4 v4.1.17 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.63.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/assert v1.3.1 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.etcd.io/bbolt v1.4.2 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache v0.43.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.62.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.62.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.37.0 // indirect
	go.opentelemetry.io/otel v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.37.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/sdk v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.0 // indirect
	go.temporal.io/api v1.50.0 // indirect
	go.temporal.io/sdk v1.34.0 // indirect
	go.temporal.io/sdk/contrib/opentelemetry v0.6.0 // indirect
	go.temporal.io/sdk/contrib/tally v0.2.0 // indirect
	go.temporal.io/server v1.28.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/mock v0.5.2 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/oauth2 v0.30.0 // indirect
	golang.org/x/sync v0.15.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/time v0.12.0 // indirect
	golang.org/x/tools v0.34.0 // indirect
	google.golang.org/api v0.239.0 // indirect
	google.golang.org/genproto v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
