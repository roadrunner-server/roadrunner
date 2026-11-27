# Changelog

releases: [docs](https://docs.roadrunner.dev/docs/releases)

## [3.2.1](https://github.com/roadrunner-server/roadrunner/compare/v3.2.0...v3.2.1) (2026-11-27)


### Bug Fixes

* **config:** load endure settings through config plugin ([bf62005](https://github.com/roadrunner-server/roadrunner/commit/bf620058b14e7d3495edfad619a3cb850e6bd566))
* **logs:** preserve in-flight logs during shutdown ([2790025](https://github.com/roadrunner-server/roadrunner/commit/279002501773750ec8aa5b7ba707a6c0037d4059))

## 3.2.0 (2026-11-26)

### 🎯 Core

- ✨ **Linux packages**: Install RoadRunner from an RPM on amd64 or arm64. Debian packages now use the name `roadrunner` across versions and set file ownership to `root`. See [RPM packages](https://github.com/roadrunner-server/roadrunner/pull/2460) and [Debian packages](https://github.com/roadrunner-server/roadrunner/pull/2467).
- ✨ **Composer package**: The v3 metapackage moved from `spiral/roadrunner` to `roadrunner/roadrunner`. See [PR #2431](https://github.com/roadrunner-server/roadrunner/pull/2431).
- 🐛 **Configuration schema**: RoadRunner uses the `logger` and `amqp` plugin schemas for v3 validation. The logger schema accepts `format` and `time_format`, rejects `file_logger_options`, and removes a trailing comma in channel options. The AMQP schema and sample configuration use named connections and nested `exchange` and `queue` settings. See [BUG #2420](https://github.com/roadrunner-server/roadrunner/issues/2420), [PR #2461](https://github.com/roadrunner-server/roadrunner/pull/2461), and [logger PR #128](https://github.com/roadrunner-server/logger/pull/128).
- 🐛 **Security and dependencies**: Update Go in `go.mod` to `1.27.3` ([CVE-2026-111943](https://pkg.go.dev/vuln/GO-2026-6809)) and `golang.org/x/net` to `v0.63.0` ([CVE-2026-78661](https://www.cve.org/CVERecord?id=CVE-2026-78661)). Retract the accidentally published `v3.1.0` Go module version. See [Issue #2456](https://github.com/roadrunner-server/roadrunner/issues/2456) and [PR #2429](https://github.com/roadrunner-server/roadrunner/pull/2429).

### 📦 `http` plugin

- ✨ **HTTP/3**: Configure `http.http3` to start the HTTP/3 listener without the `-e` flag. See [PR #294](https://github.com/roadrunner-server/http/pull/294).
- 🐛 **Multipart upload cleanup**: Remove temporary multipart files when request form parsing fails. This prevents unauthenticated temporary disk fill on HTTP/3 and FastCGI listeners. See [GHSA-f9gq-r78x-2pg4](https://github.com/roadrunner-server/roadrunner/security/advisories/GHSA-f9gq-r78x-2pg4) and [PR #314](https://github.com/roadrunner-server/http/pull/314).
- 🐛 **ACME issuer and renewal**: Apply `email`, `use_production_endpoint`, `challenge_type`, and alternate port settings to the configured CertMagic issuer and restore certificate renewal. See [BUG #2455](https://github.com/roadrunner-server/roadrunner/issues/2455) and [PR #307](https://github.com/roadrunner-server/http/pull/307).

### 📦 `fileserver` plugin

- ✨ **Fiber v3**: Use Fiber v3 for static file serving and match URL prefixes at path-segment boundaries. See [PR #107](https://github.com/roadrunner-server/fileserver/pull/107).

### 📦 `jobs` plugin

- ✨ **Redis Streams**: Use `driver: redis` to run Jobs pipelines on Redis 8.0 or later. The driver supports delayed delivery and recovers unacknowledged jobs when a consumer stops. See [FR #965](https://github.com/roadrunner-server/roadrunner/issues/965) and [PR #2466](https://github.com/roadrunner-server/roadrunner/pull/2466).

### 📦 `kafka` driver

- ✨ **Partition ordering**: Set `consumer_options.pipelining_strategy` to `Serial` to finish each record and its retries before processing the next record from that partition. Other partitions can still run in parallel. See [FR #2222](https://github.com/roadrunner-server/roadrunner/issues/2222) and [PR #632](https://github.com/roadrunner-server/kafka/pull/632).
- 🐛 **Pause and resume**: Start one consumer listener per pipeline across pause and resume operations. See [PR #632](https://github.com/roadrunner-server/kafka/pull/632).

### 📦 `redis` driver

- ✨ **Native Redis commands**: Call Redis commands over RPC using a named Redis KV connection. For example, use `redis.IncrBy` for atomic counters, `redis.HSet` for hashes, and `redis.ZAdd` for sorted sets. See [FR #1195](https://github.com/roadrunner-server/roadrunner/issues/1195) and [PR #250](https://github.com/roadrunner-server/redis/pull/250).

### 📦 `metrics` plugin

- ✨ **Prometheus metrics**: Set `metrics.prefix` and `metrics.labels` to add a name prefix and constant labels to metrics from RoadRunner plugins. Application, Go, and process metrics keep their names and labels. See [FR #2145](https://github.com/roadrunner-server/roadrunner/issues/2145) and [PR #208](https://github.com/roadrunner-server/metrics/pull/208).

### 📦 `temporal` plugin

- ✨ **Activity resets**: `temporal.RecordActivityHeartbeat` returns `"reset": true` when the server resets a running activity. With [Temporal PHP SDK 2.19](https://github.com/temporalio/sdk-php/releases/tag/v2.19) or later, `Activity::heartbeat()` throws `ActivityResetException` so PHP code can distinguish a reset from cancellation. See [FR #2249](https://github.com/roadrunner-server/roadrunner/issues/2249) and [PR #842](https://github.com/temporalio/roadrunner-temporal/pull/842).

## 3.0.0 (2026-10-08)

RoadRunner v3 uses v6 Go plugins. See [Upgrade from v2025 to v3](https://github.com/roadrunner-server/docs/blob/release/v3/intro/v3-migration.md) for the required configuration and API changes.

### 🎯 Core

- ✨ **Go module**: Import RoadRunner through `github.com/roadrunner-server/roadrunner/v3`. See [Embedding a server](https://github.com/roadrunner-server/docs/blob/release/v3/customization/embedding.md).
- ✨ **Plugin API**: Protobuf schemas, generated Go messages, and Go plugin contracts now have separate repositories: `api`, `api-go/v6`, and `api-plugins/v6`. Custom plugins use `log/slog`, `pool/v2`, and `goridge/v4`. See [Writing a Plugin](https://github.com/roadrunner-server/docs/blob/release/v3/customization/plugin.md), [PR](https://github.com/roadrunner-server/roadrunner/pull/2383).
- ✨ **Unix sockets**: Configure socket mode, owner, and group for supported listeners and worker relays. See [Configuration](https://github.com/roadrunner-server/docs/blob/release/v3/intro/config.md#unix-socket-attributes), [FR](https://github.com/roadrunner-server/roadrunner/issues/1789).
- ✨ **Environment files**: Load the root `envfile` without experimental mode. A configured file is required when the server starts. See [Environment](https://github.com/roadrunner-server/docs/blob/release/v3/php/environment.md).
- ✨ **RPC codecs**: Goridge v4 removes MessagePack support. JSON, protobuf, Gob, and raw payloads remain available through the existing Goridge RPC transport. See [RPC](https://github.com/roadrunner-server/docs/blob/release/v3/php/rpc.md), [PR](https://github.com/roadrunner-server/goridge/pull/194).
- ✨ **Worker scaling**: Retry worker acquisition after automatic scale-up and revise idle-worker removal. `dynamic_allocator.max_workers` counts workers added above the base pool size. See [Auto workers scaling](https://github.com/roadrunner-server/docs/blob/release/v3/php/auto-scaling.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2086).
- ✨ **Initialization environment**: Values in `server.on_init.env` override inherited values. See [Server](https://github.com/roadrunner-server/docs/blob/release/v3/plugins/server.md), [PR](https://github.com/roadrunner-server/server/pull/137).
- ✨ **Default plugins**: Add NSQ, HTTP rate limiting, Zstd compression, and the protobuf registry to the default container. TCP requires a [custom build](https://github.com/roadrunner-server/docs/blob/release/v3/plugins/tcp.md). See [NSQ and TCP](https://github.com/roadrunner-server/roadrunner/pull/2383), [rate limiting](https://github.com/roadrunner-server/roadrunner/pull/2389), and [Zstd and Protoreg](https://github.com/roadrunner-server/roadrunner/pull/2410).

### 📦 `http` plugin

- ✨ **Middleware order**: Requests enter `http.middleware` from left to right. Reverse a v5 list to keep its previous order. See [Middleware order](https://github.com/roadrunner-server/docs/blob/release/v3/http/http.md#middleware-order), [PR](https://github.com/roadrunner-server/http/pull/226).
- ✨ **Cleartext HTTP/2**: Use HTTP/2 prior knowledge for H2C connections. HTTP/1.1 `Upgrade: h2c` is no longer supported. See [HTTP/2](https://github.com/roadrunner-server/docs/blob/release/v3/http/http.md#http2).
- ✨ **PROXY protocol**: Accept v1 and v2 headers on HTTP and HTTPS listeners with explicit trusted proxy addresses. See [HTTP](https://github.com/roadrunner-server/docs/blob/release/v3/http/http.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/937).

### 📦 `proxy_ip_parser` middleware

- ✨ **Trusted headers**: Select forwarding headers and their priority through `http.trusted_headers`. See [Proxy IP parser](https://github.com/roadrunner-server/docs/blob/release/v3/http/proxy.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/1515).

### 📦 `static` middleware

- ✨ **File serving controls**: Configure URL prefixes, metadata cache TTL, missing-file cache TTL, and cache limits. Reset both caches with `rr reset static`. Weak ETags now use file size and modification time. See [Static files](https://github.com/roadrunner-server/docs/blob/release/v3/http/static.md), [PR](https://github.com/roadrunner-server/static/pull/147).

### 📦 `rate_limiter` middleware

- ✨ **Request limits**: Limit requests by global, IP, or header key. Rejected requests return `429` with `Retry-After`. See [Rate limiter](https://github.com/roadrunner-server/docs/blob/release/v3/http/rate-limiter.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/934).

### 📦 `zstd` middleware

- ✨ **Response compression**: Compress HTTP responses with Zstandard when the client accepts `zstd`. See [Zstd](https://github.com/roadrunner-server/docs/blob/release/v3/http/zstd.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2080).

### 📦 `grpc` plugin

- ✨ **Server reflection**: Discover services through reflection and retrieve PHP service descriptors through the bundled Protoreg plugin. See [Protoreg](https://github.com/roadrunner-server/docs/blob/release/v3/grpc/protoreg.md#server-reflection), [FR](https://github.com/roadrunner-server/roadrunner/issues/1208).
- ✨ **Error details**: Include supported `google.rpc.Status` details in failed unary-call logs. See [gRPC](https://github.com/roadrunner-server/docs/blob/release/v3/grpc/grpc.md#error-details-in-logs), [FR](https://github.com/roadrunner-server/roadrunner/issues/1897).

### 📦 `amqp` driver

- ✨ **Named connections**: Connect pipelines to different brokers. Static pipelines require `config.connection` and nested `exchange` and `queue` settings; AMQP `config.version` is removed. See [RabbitMQ](https://github.com/roadrunner-server/docs/blob/release/v3/queues/amqp.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/1962).
- ✨ **Declaration controls**: Configure queue and exchange declarations independently. Use existing broker resources with accounts that cannot declare them. See [RabbitMQ](https://github.com/roadrunner-server/docs/blob/release/v3/queues/amqp.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2275).

### 📦 `jobs` plugin

- ✨ **Named worker pools**: Assign pipelines to separate PHP worker pools with their own commands and capacity. The `pool` header selects the destination pool. See [Jobs](https://github.com/roadrunner-server/docs/blob/release/v3/queues/overview-queues.md#named-worker-pools), [PR](https://github.com/roadrunner-server/jobs/pull/139).
- ✨ **Trace continuity**: Preserve valid producer trace context when jobs are submitted through RPC. See [Jobs](https://github.com/roadrunner-server/docs/blob/release/v3/queues/overview-queues.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2211).
- ✨ **Outcome metrics**: Count successful, failed, and requeued attempts separately with `rr_jobs_jobs_requeue`. Outcome and push totals use counters. See [Metrics](https://github.com/roadrunner-server/docs/blob/release/v3/lab/metrics.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/1566).

### 📦 `nsq` driver

- ✨ **NSQ jobs**: Publish and consume jobs through NSQ topics and channels. Support broker discovery, delayed delivery, acknowledgements, and configurable reconnection intervals. See [NSQ](https://github.com/roadrunner-server/docs/blob/release/v3/queues/nsq.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/1948).

### 📦 `nats` driver

- ✨ **Stream retention**: Preserve stream contents when a pipeline stops unless stream deletion is enabled. Review retention and repeated-delivery handling when upgrading. See [NATS](https://github.com/roadrunner-server/docs/blob/release/v3/queues/nats.md), [PR](https://github.com/roadrunner-server/nats/pull/226).
- ✨ **Acknowledgment wait**: Configure `ack_wait` for JetStream redelivery timing. See [NATS](https://github.com/roadrunner-server/docs/blob/release/v3/queues/nats.md), [PR](https://github.com/roadrunner-server/nats/pull/190).

### 📦 `kafka` driver

- ✨ **Direct partition consumption**: Apply `consume_partitions` to select topic partitions and starting offsets. See [Kafka](https://github.com/roadrunner-server/docs/blob/release/v3/queues/kafka.md), [PR](https://github.com/roadrunner-server/kafka/pull/580).
- 🐛 **Requeue isolation**: A requeued job, or a job nacked with redelivery, goes back to the RoadRunner priority queue. The driver no longer produces a copy to the source topic, so other consumer groups do not receive the retry. The driver honors the requested delay. A burst of retries no longer drops jobs. A retry waits in memory: it is lost after a stop, restart, or crash once a later offset of its partition is committed. See [Kafka](https://github.com/roadrunner-server/docs/blob/release/v3/queues/kafka.md), [BUG](https://github.com/roadrunner-server/roadrunner/issues/2413).
- 🐛 **Consumer recovery**: When the consumer gets a non-retriable Kafka error, for example when the broker denies access to the topic or removes the consumer from the group, the `jobs` plugin restarts the pipeline. The old consumer closes, and a pipeline in `jobs.consume` continues with a new consumer. If the restart of a configured pipeline fails, RoadRunner stops with an error. See [Kafka](https://github.com/roadrunner-server/docs/blob/release/v3/queues/kafka.md), [BUG](https://github.com/roadrunner-server/roadrunner/issues/2391).

### 📦 `sqs` driver

- 🐛 **FIFO retry delivery**: Use a fresh broker deduplication ID for each republished retry while retaining the application job ID. This prevents SQS from discarding the retry as a duplicate of the deleted original. See [SQS](https://github.com/roadrunner-server/docs/blob/release/v3/queues/sqs.md), [PR](https://github.com/roadrunner-server/sqs/pull/864).

### 📦 `beanstalk` driver

- ✨ **Job headers**: Preserve headers through storage and retries, including trace context and named-pool routing. Existing messages cannot recover headers that were not stored. See [Beanstalk](https://github.com/roadrunner-server/docs/blob/release/v3/queues/beanstalk.md).

### 📦 `logger` plugin

- ✨ **Logging configuration**: Use Go `log/slog` with JSON, text, and raw formats. Configure file destinations through `output`; built-in rotation and `file_logger_options` are removed. See [Logger](https://github.com/roadrunner-server/docs/blob/release/v3/lab/logger.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/1750).
- ✨ **Custom formats**: Set `format` placeholders and `time_format` to control log records and timestamps. See [Logger](https://github.com/roadrunner-server/docs/blob/release/v3/lab/logger.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/1376).
- ✨ **Non-blocking output**: Queue log records before writing them. A full output queue drops new records to keep application calls responsive. See [Logger](https://github.com/roadrunner-server/docs/blob/release/v3/lab/logger.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2195).

### 📦 `lock` plugin

- ✨ **Redis backend**: Share locks between RoadRunner instances through Redis with the existing lock RPC API. The in-memory backend remains the default. See [Locks](https://github.com/roadrunner-server/docs/blob/release/v3/plugins/locks.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2070).

### 📦 `redis` driver

- ✨ **Sentinel authentication**: Configure `sentinel_username` and `sentinel_password` independently from Redis master credentials. Password-only Sentinel authentication remains supported. See [Redis](https://github.com/roadrunner-server/docs/blob/release/v3/kv/redis.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2407).

### 📦 `service` plugin

- ✨ **Deferred updates**: Change stored service configuration through `service.Update`. Existing processes keep their current settings; new processes use the updated values. PHP applications need client and DTO support for this method. See [Service](https://github.com/roadrunner-server/docs/blob/release/v3/plugins/service.md), [FR](https://github.com/roadrunner-server/roadrunner/issues/2119).

### 📦 `status` plugin

- ✨ **Kubernetes probes**: Use `/livez` as an alias for `/health` and `/readyz` as an alias for `/ready`. See [HealthChecks](https://github.com/roadrunner-server/docs/blob/release/v3/lab/health.md), [PR](https://github.com/roadrunner-server/status/pull/154).

### 📦 `otel` plugin

- ✨ **Span timing**: End named middleware spans before the next handler starts. Use the outer HTTP server span for request latency. See [OpenTelemetry](https://github.com/roadrunner-server/docs/blob/release/v3/lab/otel.md) and [HTTP tracing](https://github.com/roadrunner-server/docs/blob/release/v3/http/http.md#tracing), [FR](https://github.com/roadrunner-server/roadrunner/issues/2276).
- ✨ **Exporters**: Remove the native Zipkin exporter. Send traces through OTLP to a compatible receiver. See [OpenTelemetry](https://github.com/roadrunner-server/docs/blob/release/v3/lab/otel.md).

### 📦 `centrifuge` plugin

- ✨ **Protocol update**: Add `NotifyCacheEmpty` event support and remove `centrifuge.RateLimit`. The new event needs matching PHP DTO and handler support. See [Centrifuge](https://github.com/roadrunner-server/docs/blob/release/v3/plugins/centrifuge.md).

### 📦 `temporal` plugin

- ✨ **Worker heartbeats**: Configure the worker heartbeat interval and report PHP SDK identity with host CPU and memory metrics. See [Temporal](https://github.com/roadrunner-server/docs/blob/release/v3/workflow/temporal.md), [PR](https://github.com/temporalio/roadrunner-temporal/pull/790).
- ✨ **Dynamic workflows**: Register a catch-all workflow for workflow types without a named registration. See [Worker](https://github.com/roadrunner-server/docs/blob/release/v3/workflow/worker.md), [PR](https://github.com/temporalio/roadrunner-temporal/pull/784).
- 🐛 **Activity-worker recovery**: Replace a failed activity worker without restarting healthy workers in the same activity pool. See [Worker](https://github.com/roadrunner-server/docs/blob/release/v3/workflow/worker.md), [BUG](https://github.com/roadrunner-server/roadrunner/issues/2335).

### 📦 `velox` builder

- ✨ **Module configuration**: Select plugins through `[plugins.<name>]`, with `module_name` and `tag`. Support module replacements, exclusions, and repeatable build timestamps. The remote build server and Windows targets are removed. See [Building a Server](https://github.com/roadrunner-server/docs/blob/release/v3/customization/build.md), [PR](https://github.com/roadrunner-server/velox/pull/350).
