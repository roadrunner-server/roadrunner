module github.com/roadrunner-server/roadrunner/v2024

go 1.22.4

require (
	github.com/buger/goterm v1.0.4
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.17.0
	github.com/joho/godotenv v1.5.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/roadrunner-server/amqp/v4 v4.11.3
	github.com/roadrunner-server/api/v4 v4.12.0
	github.com/roadrunner-server/app-logger/v4 v4.4.3
	github.com/roadrunner-server/beanstalk/v4 v4.8.3
	github.com/roadrunner-server/boltdb/v4 v4.9.3
	github.com/roadrunner-server/centrifuge/v4 v4.9.3
	github.com/roadrunner-server/config/v4 v4.9.1
	github.com/roadrunner-server/endure/v2 v2.4.5
	github.com/roadrunner-server/errors v1.4.0
	github.com/roadrunner-server/fileserver/v4 v4.3.3
	github.com/roadrunner-server/goridge/v3 v3.8.2
	github.com/roadrunner-server/grpc/v4 v4.8.5
	github.com/roadrunner-server/gzip/v4 v4.3.3
	github.com/roadrunner-server/headers/v4 v4.4.3
	github.com/roadrunner-server/http/v4 v4.7.5
	github.com/roadrunner-server/informer/v4 v4.5.3
	github.com/roadrunner-server/jobs/v4 v4.9.3
	github.com/roadrunner-server/kafka/v4 v4.6.3
	github.com/roadrunner-server/kv/v4 v4.6.3
	github.com/roadrunner-server/lock/v4 v4.7.3
	github.com/roadrunner-server/logger/v4 v4.4.3
	github.com/roadrunner-server/memcached/v4 v4.5.3
	github.com/roadrunner-server/memory/v4 v4.8.3
	github.com/roadrunner-server/metrics/v4 v4.3.3
	github.com/roadrunner-server/nats/v4 v4.8.3
	github.com/roadrunner-server/otel/v4 v4.5.3
	github.com/roadrunner-server/prometheus/v4 v4.3.3
	github.com/roadrunner-server/proxy_ip_parser/v4 v4.3.3
	github.com/roadrunner-server/redis/v4 v4.4.3
	github.com/roadrunner-server/resetter/v4 v4.3.3
	github.com/roadrunner-server/rpc/v4 v4.4.3
	github.com/roadrunner-server/sdk/v4 v4.7.3
	github.com/roadrunner-server/send/v4 v4.4.3
	github.com/roadrunner-server/server/v4 v4.8.3
	github.com/roadrunner-server/service/v4 v4.7.3
	github.com/roadrunner-server/sqs/v4 v4.8.3
	github.com/roadrunner-server/static/v4 v4.3.3
	github.com/roadrunner-server/status/v4 v4.6.3
	github.com/roadrunner-server/tcp/v4 v4.5.3
	github.com/spf13/cobra v1.8.0
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.9.0
	github.com/temporalio/roadrunner-temporal/v4 v4.8.2
	go.uber.org/automaxprocs v1.5.3
)

replace github.com/uber-go/tally/v4 => github.com/uber-go/tally/v4 v4.1.10

exclude (
	github.com/spf13/viper v1.18.0
	github.com/spf13/viper v1.18.1
	go.temporal.io/api v1.26.1
)

require (
	github.com/andybalholm/brotli v1.1.0 // indirect
	github.com/aws/aws-sdk-go v1.53.17 // indirect
	github.com/aws/aws-sdk-go-v2 v1.27.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.17 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.17 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.8 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.32.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.24.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.11 // indirect
	github.com/aws/smithy-go v1.20.2 // indirect
	github.com/beanstalkd/go-beanstalk v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874 // indirect
	github.com/cactus/go-statsd-client/v5 v5.1.0 // indirect
	github.com/caddyserver/certmagic v0.21.3 // indirect
	github.com/caddyserver/zerossl v0.1.3 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emicklei/proto v1.13.2 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fsnotify/fsnotify v1.7.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/gofiber/fiber/v2 v2.52.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/google/pprof v0.0.0-20240528025155-186aa0362fba // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.20.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/klauspost/compress v1.17.8 // indirect
	github.com/klauspost/cpuid/v2 v2.2.7 // indirect
	github.com/libdns/libdns v0.2.2 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/mholt/acmez v1.2.0 // indirect
	github.com/mholt/acmez/v2 v2.0.1 // indirect
	github.com/miekg/dns v1.1.59 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nats.go v1.35.0 // indirect
	github.com/nats-io/nkeys v0.4.7 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/ginkgo/v2 v2.19.0 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pelletier/go-toml/v2 v2.2.2 // indirect
	github.com/pierrec/lz4/v4 v4.1.21 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_golang v1.19.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.54.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/quic-go/qpack v0.4.0 // indirect
	github.com/quic-go/quic-go v0.45.0 // indirect
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.0.5 // indirect
	github.com/redis/go-redis/extra/redisotel/v9 v9.0.5 // indirect
	github.com/redis/go-redis/v9 v9.5.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/roadrunner-server/tcplisten v1.4.0 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/cors v1.11.0 // indirect
	github.com/sagikazarmark/locafero v0.6.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.6.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.14 // indirect
	github.com/tklauser/numcpus v0.8.0 // indirect
	github.com/twmb/franz-go v1.17.0 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.8.0 // indirect
	github.com/twmb/murmur3 v1.1.8 // indirect
	github.com/uber-go/tally/v4 v4.1.17-0.20240412215630-22fe011f5ff0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.54.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/blake3 v0.2.3 // indirect
	go.etcd.io/bbolt v1.3.10 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache v0.43.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.52.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.52.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.27.0 // indirect
	go.opentelemetry.io/otel v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.27.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.27.0 // indirect
	go.opentelemetry.io/otel/metric v1.27.0 // indirect
	go.opentelemetry.io/otel/sdk v1.27.0 // indirect
	go.opentelemetry.io/otel/trace v1.27.0 // indirect
	go.opentelemetry.io/proto/otlp v1.2.0 // indirect
	go.temporal.io/api v1.34.0 // indirect
	go.temporal.io/sdk v1.26.1 // indirect
	go.temporal.io/sdk/contrib/opentelemetry v0.5.0 // indirect
	go.temporal.io/sdk/contrib/tally v0.2.0 // indirect
	go.temporal.io/server v1.24.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/mock v0.4.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.24.0 // indirect
	golang.org/x/exp v0.0.0-20240604190554-fc45aab8b7f8 // indirect
	golang.org/x/mod v0.18.0 // indirect
	golang.org/x/net v0.26.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.21.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.22.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240604185151-ef581f913117 // indirect
	google.golang.org/grpc v1.64.0 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
