module github.com/roadrunner-server/roadrunner/v2023

go 1.20

require (
	github.com/buger/goterm v1.0.4
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.14.1
	github.com/joho/godotenv v1.5.1
	github.com/olekukonko/tablewriter v0.0.5
	github.com/roadrunner-server/amqp/v4 v4.1.1
	github.com/roadrunner-server/api/v3 v3.2.1
	github.com/roadrunner-server/app-logger/v4 v4.0.1
	github.com/roadrunner-server/beanstalk/v4 v4.1.1
	github.com/roadrunner-server/boltdb/v4 v4.2.1
	github.com/roadrunner-server/centrifuge/v4 v4.0.2
	github.com/roadrunner-server/config/v4 v4.1.1
	github.com/roadrunner-server/endure/v2 v2.2.0
	github.com/roadrunner-server/errors v1.2.0
	github.com/roadrunner-server/fileserver/v4 v4.0.1
	github.com/roadrunner-server/goridge/v3 v3.6.2
	github.com/roadrunner-server/grpc/v4 v4.1.0
	github.com/roadrunner-server/gzip/v4 v4.0.1
	github.com/roadrunner-server/headers/v4 v4.0.1
	github.com/roadrunner-server/http/v4 v4.0.3
	github.com/roadrunner-server/informer/v4 v4.0.1
	github.com/roadrunner-server/jobs/v4 v4.2.1
	github.com/roadrunner-server/kv/v4 v4.1.1
	github.com/roadrunner-server/logger/v4 v4.0.1
	github.com/roadrunner-server/memcached/v4 v4.1.2
	github.com/roadrunner-server/memory/v4 v4.1.1
	github.com/roadrunner-server/metrics/v4 v4.0.1
	github.com/roadrunner-server/nats/v4 v4.1.1
	github.com/roadrunner-server/otel/v4 v4.1.0
	github.com/roadrunner-server/prometheus/v4 v4.0.2
	github.com/roadrunner-server/proxy_ip_parser/v4 v4.0.2
	github.com/roadrunner-server/redis/v4 v4.1.1
	github.com/roadrunner-server/reload/v4 v4.0.1
	github.com/roadrunner-server/resetter/v4 v4.0.1
	github.com/roadrunner-server/rpc/v4 v4.1.0
	github.com/roadrunner-server/sdk/v4 v4.0.0
	github.com/roadrunner-server/send/v4 v4.0.2
	github.com/roadrunner-server/server/v4 v4.0.1
	github.com/roadrunner-server/service/v4 v4.0.2
	github.com/roadrunner-server/sqs/v4 v4.1.1
	github.com/roadrunner-server/static/v4 v4.0.2
	github.com/roadrunner-server/status/v4 v4.0.1
	github.com/roadrunner-server/tcp/v4 v4.0.1
	github.com/spf13/cobra v1.6.1
	github.com/spf13/viper v1.15.0
	github.com/stretchr/testify v1.8.1
	github.com/temporalio/roadrunner-temporal/v4 v4.1.0
	go.buf.build/protocolbuffers/go/roadrunner-server/api v1.3.31
	go.uber.org/automaxprocs v1.5.1
	golang.org/x/exp v0.0.0-20230213192124-5e25df0256eb
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2 v1.17.4 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.18.12 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.12 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.12.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.28 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.22 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.29 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.22 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.20.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.3 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
	github.com/beanstalkd/go-beanstalk v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20230124162541-5f7a7d875746 // indirect
	github.com/cactus/go-statsd-client/statsd v0.0.0-20200423205355-cb0885a1018c // indirect
	github.com/caddyserver/certmagic v0.17.2 // indirect
	github.com/cenkalti/backoff/v4 v4.2.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/emicklei/proto v1.11.1 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	github.com/goccy/go-json v0.10.0 // indirect
	github.com/gofiber/fiber/v2 v2.42.0 // indirect
	github.com/gogo/googleapis v1.4.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/gogo/status v1.1.1 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.15.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.15.15 // indirect
	github.com/klauspost/cpuid/v2 v2.2.3 // indirect
	github.com/libdns/libdns v0.2.1 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/mholt/acmez v1.1.0 // indirect
	github.com/miekg/dns v1.1.50 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/jwt/v2 v2.3.0 // indirect
	github.com/nats-io/nats.go v1.23.0 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/onsi/gomega v1.24.1 // indirect
	github.com/openzipkin/zipkin-go v0.4.1 // indirect
	github.com/pborman/uuid v1.2.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/philhofer/fwd v1.1.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.39.0 // indirect
	github.com/prometheus/procfs v0.9.0 // indirect
	github.com/rabbitmq/amqp091-go v1.7.0 // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	github.com/roadrunner-server/api/v4 v4.1.0 // indirect
	github.com/roadrunner-server/tcplisten v1.2.1 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/savsgio/dictpool v0.0.0-20221023140959-7bf2e61cea94 // indirect
	github.com/savsgio/gotils v0.0.0-20230208104028-c358bd845dee // indirect
	github.com/segmentio/asm v1.2.0 // indirect
	github.com/segmentio/encoding v0.3.6 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spf13/afero v1.9.3 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.4.2 // indirect
	github.com/tinylib/msgp v1.1.8 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/twmb/murmur3 v1.1.6 // indirect
	github.com/uber-go/tally/v4 v4.1.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.44.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.buf.build/grpc/go/roadrunner-server/api v1.4.31 // indirect
	go.etcd.io/bbolt v1.3.7 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.39.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.39.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.14.0 // indirect
	go.opentelemetry.io/otel v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/jaeger v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/internal/retry v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.13.0 // indirect
	go.opentelemetry.io/otel/exporters/zipkin v1.13.0 // indirect
	go.opentelemetry.io/otel/metric v0.36.0 // indirect
	go.opentelemetry.io/otel/sdk v1.13.0 // indirect
	go.opentelemetry.io/otel/trace v1.13.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	go.temporal.io/api v1.17.0 // indirect
	go.temporal.io/sdk v1.21.1 // indirect
	go.temporal.io/sdk/contrib/opentelemetry v0.2.0 // indirect
	go.temporal.io/sdk/contrib/tally v0.2.0 // indirect
	go.temporal.io/server v1.19.1 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/net v0.6.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	google.golang.org/genproto v0.0.0-20230209215440-0dfe4f8abfcc // indirect
	google.golang.org/grpc v1.53.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
