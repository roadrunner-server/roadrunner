module github.com/spiral/roadrunner/v2

go 1.16

require (
	github.com/Shopify/toxiproxy v2.1.4+incompatible
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/alicebob/miniredis/v2 v2.15.1
	// ========= AWS SDK v2
	github.com/aws/aws-sdk-go-v2 v1.8.0
	github.com/aws/aws-sdk-go-v2/config v1.6.0
	github.com/aws/aws-sdk-go-v2/credentials v1.3.2
	github.com/aws/aws-sdk-go-v2/service/sqs v1.7.1
	github.com/aws/smithy-go v1.7.0
	// =====================
	github.com/beanstalkd/go-beanstalk v0.1.0
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/fasthttp/websocket v1.4.3
	github.com/fatih/color v1.12.0
	github.com/go-redis/redis/v8 v8.11.3
	github.com/gofiber/fiber/v2 v2.17.0
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.3.0
	github.com/json-iterator/go v1.1.11
	github.com/klauspost/compress v1.13.4
	github.com/prometheus/client_golang v1.11.0
	github.com/rabbitmq/amqp091-go v0.0.0-20210812094702-b2a427eb7d17
	github.com/shirou/gopsutil v3.21.7+incompatible
	github.com/spf13/viper v1.8.1
	// SPIRAL ====
	github.com/spiral/endure v1.0.2
	github.com/spiral/errors v1.0.11
	github.com/spiral/goridge/v3 v3.2.0
	// ===========
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.7 // indirect
	github.com/valyala/tcplisten v1.0.0
	github.com/yookoala/gofast v0.6.0
	go.etcd.io/bbolt v1.3.6
	go.uber.org/multierr v1.7.0
	go.uber.org/zap v1.19.0
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e
	google.golang.org/protobuf v1.27.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
)
