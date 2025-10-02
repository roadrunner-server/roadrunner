# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

RoadRunner is a high-performance PHP application server and process manager written in Go. It supports running as a service with extensive plugin functionality for HTTP/2/3, gRPC, queues (RabbitMQ, Kafka, SQS, NATS), KV stores, WebSockets, Temporal workflows, and more.

## Development Commands

### Build
```bash
make build
# Or manually:
CGO_ENABLED=0 go build -trimpath -ldflags "-s" -o rr cmd/rr/main.go
```

### Test
```bash
make test
# Or manually:
go test -v -race ./...
```

### Debug
```bash
make debug
# Uses delve to debug with sample config
```

### Run RoadRunner
```bash
./rr serve -c .rr.yaml
```

### Other Commands
```bash
./rr workers          # Show worker status
./rr workers -i       # Interactive worker information
./rr reset            # Reset workers
./rr jobs             # Jobs management commands
./rr stop             # Stop RoadRunner server
```

### Run Single Test
```bash
go test -v -race -run TestName ./path/to/package
```

## Architecture

### Plugin System

RoadRunner uses the **Endure** dependency injection container. All plugins are registered in `container/plugins.go:Plugins()`. The plugin architecture follows these principles:

1. **Plugin Registration**: Plugins are listed in `container/plugins.go` and automatically wired by Endure
2. **Plugin Dependencies**: Plugins declare dependencies via struct fields with interface types
3. **Initialization Order**: Endure resolves the dependency graph and initializes plugins in correct order

### Key Components

- **`cmd/rr/main.go`**: Entry point that delegates to CLI commands
- **`internal/cli/`**: CLI command implementations (serve, workers, reset, jobs, stop)
- **`container/`**: Plugin registration and Endure container configuration
- **Plugin packages**: External packages under `github.com/roadrunner-server/*` (imported in go.mod)

### Configuration

- Primary config: `.rr.yaml` (extensive sample provided)
- Version 3 config format required (`version: '3'`)
- Environment variable substitution supported: `${ENVIRONMENT_VARIABLE_NAME}`
- Sample configs: `.rr-sample-*.yaml` for different use cases (HTTP, gRPC, Temporal, Kafka, etc.)

### Core Plugins

**Server Management:**
- `server`: Worker pool management (NewWorker, NewWorkerPool)
- `rpc`: RPC server for PHP-to-Go communication (default: tcp://127.0.0.1:6001)
- `logger`: Logging infrastructure
- `informer`: Worker status reporting
- `resetter`: Worker reset functionality

**Protocol Servers:**
- `http`: HTTP/1/2/3 and FastCGI server with middleware support
- `grpc`: gRPC server
- `tcp`: Raw TCP connection handling

**Jobs/Queue Drivers:**
- `jobs`: Core jobs plugin
- `amqp`, `sqs`, `nats`, `kafka`, `beanstalk`: Queue backends
- `gps`: Google Pub/Sub

**KV Stores:**
- `kv`: Core KV plugin
- `memory`, `boltdb`, `redis`, `memcached`: Storage backends

**HTTP Middleware:**
- `static`, `headers`, `gzip`, `prometheus`, `send`, `proxy_ip_parser`, `otel`, `fileserver`

**Other:**
- `temporal`: Temporal.io workflow engine integration
- `centrifuge`: WebSocket/Broadcast via Centrifugo
- `lock`: Distributed locks
- `metrics`: Prometheus metrics
- `service`: Systemd-like service manager

### Worker Communication

RoadRunner communicates with PHP workers via:
- **Goridge protocol**: Binary protocol over pipes, TCP, or Unix sockets
- **RPC**: For management operations (reset, stats, etc.)
- Workers are PHP processes that implement the RoadRunner worker protocol

### Testing

- Tests use standard Go testing with `-race` flag
- Test files follow `*_test.go` convention
- Sample configs in `.rr-sample-*.yaml` are used for integration tests
- Test directories: `container/test`, `internal/rpc/test`

## Important Notes

- Go version: 1.25+ required (see go.mod)
- Module path: `github.com/roadrunner-server/roadrunner/v2025`
- Some versions are explicitly excluded in go.mod (e.g., go-redis v9.15.0, viper v1.18.x)
- Debug mode available via `--debug` flag (starts debug server on :6061)
- Config overrides supported via `-o dot.notation=value` flag
- Working directory can be set with `-w` flag
- `.env` file support via `--dotenv` flag or `DOTENV_PATH` environment variable

## Adding New Plugins

1. Import the plugin package in `container/plugins.go`
2. Add plugin instance to the `Plugins()` slice
3. Plugin must implement appropriate RoadRunner plugin interfaces
4. Endure will handle dependency injection and lifecycle management

## Configuration Patterns

- Each plugin has its own configuration section (named after plugin)
- Pools configuration is consistent across plugins (num_workers, max_jobs, timeouts, supervisor)
- TLS configuration follows similar pattern across plugins
- Most plugins support graceful shutdown via timeouts
