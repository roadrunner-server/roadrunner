module github.com/roadrunner-server/roadrunner/v2025

go 1.27

require (
	github.com/buger/goterm v1.0.4
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.19.0
	github.com/joho/godotenv v1.5.1
	github.com/olekukonko/tablewriter v1.1.4
	github.com/roadrunner-server/amqp/v6 v6.0.0-beta.8
	github.com/roadrunner-server/api-go/v6 v6.0.0-beta.14
	github.com/roadrunner-server/api-plugins/v6 v6.0.0-beta.2
	github.com/roadrunner-server/app-logger/v6 v6.0.0-beta.5
	github.com/roadrunner-server/beanstalk/v6 v6.0.0-beta.4
	github.com/roadrunner-server/boltdb/v6 v6.0.0-beta.4
	github.com/roadrunner-server/centrifuge/v6 v6.0.0-beta.4
	github.com/roadrunner-server/config/v6 v6.0.0-beta.3
	github.com/roadrunner-server/endure/v2 v2.6.2
	github.com/roadrunner-server/errors v1.5.0
	github.com/roadrunner-server/fileserver/v6 v6.0.0-beta.5
	github.com/roadrunner-server/google-pub-sub/v6 v6.0.0-beta.3
	github.com/roadrunner-server/goridge/v4 v4.0.0-beta.3
	github.com/roadrunner-server/grpc/v6 v6.0.0-beta.5
	github.com/roadrunner-server/gzip/v6 v6.0.0-beta.2
	github.com/roadrunner-server/headers/v6 v6.0.0-beta.2
	github.com/roadrunner-server/http/v6 v6.0.0-beta.9
	github.com/roadrunner-server/informer/v6 v6.0.0-beta.4
	github.com/roadrunner-server/jobs/v6 v6.0.0-beta.9
	github.com/roadrunner-server/kafka/v6 v6.0.0-beta.6
	github.com/roadrunner-server/kv/v6 v6.0.0-beta.7
	github.com/roadrunner-server/lock/v6 v6.0.0-beta.6
	github.com/roadrunner-server/logger/v6 v6.0.0-beta.3
	github.com/roadrunner-server/memcached/v6 v6.0.0-beta.4
	github.com/roadrunner-server/memory/v6 v6.0.0-beta.4
	github.com/roadrunner-server/metrics/v6 v6.0.0-beta.6
	github.com/roadrunner-server/nats/v6 v6.0.0-beta.4
	github.com/roadrunner-server/nsq/v6 v6.0.0-20260818191817-3a968038c195
	github.com/roadrunner-server/otel/v6 v6.0.0-beta.4.0.20260821060711-7bdd363e40b0
	github.com/roadrunner-server/pool/v2 v2.0.0-beta.1
	github.com/roadrunner-server/prometheus/v6 v6.0.0-beta.2
	github.com/roadrunner-server/proxy_ip_parser/v6 v6.0.0-beta.3
	github.com/roadrunner-server/redis/v6 v6.0.0-beta.4
	github.com/roadrunner-server/resetter/v6 v6.0.0-beta.5
	github.com/roadrunner-server/rpc/v6 v6.0.0-beta.5
	github.com/roadrunner-server/send/v6 v6.0.0-beta.4
	github.com/roadrunner-server/server/v6 v6.0.0-beta.6
	github.com/roadrunner-server/service/v6 v6.0.0-beta.7
	github.com/roadrunner-server/sqs/v6 v6.0.0-beta.5
	github.com/roadrunner-server/static/v6 v6.0.0-beta.4
	github.com/roadrunner-server/status/v6 v6.0.0-beta.7
	github.com/spf13/cobra v1.10.2
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.12.1
)

require (
	cloud.google.com/go v0.123.0 // indirect
	cloud.google.com/go/auth v0.20.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	cloud.google.com/go/iam v1.13.0 // indirect
	cloud.google.com/go/pubsub/v2 v2.6.0 // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.41.7 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.32.17 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.19.16 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.9 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.23 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.11 // indirect
	github.com/aws/aws-sdk-go-v2/service/sqs v1.42.27 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.21 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.42.1 // indirect
	github.com/aws/smithy-go v1.25.1 // indirect
	github.com/beanstalkd/go-beanstalk v0.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/bradfitz/gomemcache v0.0.0-20260422231931-4d751bb6e37c // indirect
	github.com/caddyserver/certmagic v0.25.4 // indirect
	github.com/caddyserver/zerossl v0.1.5 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clipperhouse/displaywidth v0.10.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/emicklei/proto v1.14.3 // indirect
	github.com/facebookgo/clock v0.0.0-20150410010913-600d898af40a // indirect
	github.com/felixge/httpsnoop v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/gofiber/fiber/v2 v2.52.13 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.7.0-rc.1 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.17 // indirect
	github.com/googleapis/gax-go/v2 v2.23.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.30.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/klauspost/compress v1.19.1 // indirect
	github.com/klauspost/cpuid/v2 v2.4.0 // indirect
	github.com/libdns/libdns v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.23 // indirect
	github.com/mholt/acmez v1.2.0 // indirect
	github.com/mholt/acmez/v3 v3.1.6 // indirect
	github.com/miekg/dns v1.1.72 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/nats.go v1.52.0 // indirect
	github.com/nats-io/nkeys v0.4.15 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/nexus-rpc/nexus-proto-annotations v0.1.0 // indirect
	github.com/nexus-rpc/sdk-go v0.7.0 // indirect
	github.com/nsqio/go-nsq v1.1.0 // indirect
	github.com/olekukonko/cat v0.0.0-20250911104152-50322a0618f6 // indirect
	github.com/olekukonko/errors v1.2.0 // indirect
	github.com/olekukonko/ll v0.1.6 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/quic-go/quic-go v0.61.0 // indirect
	github.com/rabbitmq/amqp091-go v1.11.0 // indirect
	github.com/redis/go-redis/extra/rediscmd/v9 v9.19.0 // indirect
	github.com/redis/go-redis/extra/redisotel/v9 v9.19.0 // indirect
	github.com/redis/go-redis/extra/redisprometheus/v9 v9.19.0 // indirect
	github.com/redis/go-redis/v9 v9.19.0 // indirect
	github.com/roadrunner-server/context v1.3.0 // indirect
	github.com/roadrunner-server/events v1.0.1 // indirect
	github.com/roadrunner-server/priority_queue v1.0.6 // indirect
	github.com/roadrunner-server/tcplisten v1.5.2 // indirect
	github.com/robfig/cron v1.2.0 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.3 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.4.0 // indirect
	github.com/tklauser/numcpus v0.12.0 // indirect
	github.com/twmb/franz-go v1.21.1 // indirect
	github.com/twmb/franz-go/pkg/kmsg v1.13.1 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.71.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	github.com/zeebo/assert v1.3.0 // indirect
	github.com/zeebo/blake3 v0.2.4 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/contrib/instrumentation/github.com/bradfitz/gomemcache/memcache/otelmemcache v0.43.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.70.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.70.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.45.0 // indirect
	go.opentelemetry.io/otel v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.45.0 // indirect
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.45.0 // indirect
	go.opentelemetry.io/otel/metric v1.45.0 // indirect
	go.opentelemetry.io/otel/sdk v1.45.0 // indirect
	go.opentelemetry.io/otel/trace v1.45.0 // indirect
	go.opentelemetry.io/proto/otlp v1.11.0 // indirect
	go.temporal.io/api v1.63.5 // indirect
	go.temporal.io/sdk v1.48.0 // indirect
	go.temporal.io/sdk/contrib/opentelemetry v0.8.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.28.0 // indirect
	go.uber.org/zap/exp v0.3.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/crypto v0.55.0 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/net v0.58.0 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/time v0.15.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	google.golang.org/api v0.287.1 // indirect
	google.golang.org/genproto v0.0.0-20260810153831-ec0a7760b754 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260819154853-08b0e4226688 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260819154853-08b0e4226688 // indirect
	google.golang.org/grpc v1.83.1 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
)

exclude (
	github.com/olekukonko/tablewriter v1.1.1
	github.com/redis/go-redis/v9 v9.15.0
	github.com/redis/go-redis/v9 v9.15.1
	github.com/spf13/viper v1.18.0
	github.com/spf13/viper v1.18.1
	go.temporal.io/api v1.26.1
)
