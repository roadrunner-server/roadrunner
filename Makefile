#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

.DEFAULT_GOAL := build

# This will output the help for each task. thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Show this help
	@printf "\033[33m%s:\033[0m\n" 'Available commands'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[32m%-14s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build RR binary file for local os/arch
	CGO_ENABLED=0 go build -trimpath -ldflags "-s" -o ./rr ./cmd/main.go

clean: ## Make some clean
	rm ./rr

install: build ## Build and install RR locally
	cp rr /usr/local/bin/rr

uninstall: ## Uninstall locally installed RR
	rm -f /usr/local/bin/rr

test_coverage:
	docker-compose -f tests/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage
	mkdir coverage
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/pipe.out -covermode=atomic ./pkg/transport/pipe
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/socket.out -covermode=atomic ./pkg/transport/socket
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/pool.out -covermode=atomic ./pkg/pool
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/worker.out -covermode=atomic ./pkg/worker
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/worker_stack.out -covermode=atomic ./pkg/worker_watcher
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/http.out -covermode=atomic ./tests/plugins/http
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/informer.out -covermode=atomic ./tests/plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/reload.out -covermode=atomic ./tests/plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/server.out -covermode=atomic ./tests/plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/status.out -covermode=atomic ./tests/plugins/status
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/config.out -covermode=atomic ./tests/plugins/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/gzip.out -covermode=atomic ./tests/plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/headers.out -covermode=atomic ./tests/plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/logger.out -covermode=atomic ./tests/plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/metrics.out -covermode=atomic ./tests/plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/redis.out -covermode=atomic ./tests/plugins/redis
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/resetter.out -covermode=atomic ./tests/plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/rpc.out -covermode=atomic ./tests/plugins/rpc
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/static.out -covermode=atomic ./tests/plugins/static
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/boltdb_unit.out -covermode=atomic ./plugins/kv/boltdb
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/kv_unit.out -covermode=atomic ./plugins/kv/memory
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/http_config.out -covermode=atomic ./plugins/http/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/memcached_unit.out -covermode=atomic ./plugins/kv/memcached
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/boltdb.out -covermode=atomic ./tests/plugins/kv/boltdb
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/memory.out -covermode=atomic ./tests/plugins/kv/memory
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/memcached.out -covermode=atomic ./tests/plugins/kv/memcached
	cat ./coverage/*.out > ./coverage/summary.out
	docker-compose -f tests/docker-compose.yaml down

test: ## Run application tests
	docker-compose -f tests/docker-compose.yaml up -d
	go test -v -race -tags=debug ./pkg/transport/pipe
	go test -v -race -tags=debug ./pkg/transport/socket
	go test -v -race -tags=debug ./pkg/pool
	go test -v -race -tags=debug ./pkg/worker
	go test -v -race -tags=debug ./pkg/worker_watcher
	go test -v -race -tags=debug ./tests/plugins/http
	go test -v -race -tags=debug ./plugins/http/config
	go test -v -race -tags=debug ./tests/plugins/informer
	go test -v -race -tags=debug ./tests/plugins/reload
	go test -v -race -tags=debug ./tests/plugins/server
	go test -v -race -tags=debug ./tests/plugins/status
	go test -v -race -tags=debug ./tests/plugins/config
	go test -v -race -tags=debug ./tests/plugins/gzip
	go test -v -race -tags=debug ./tests/plugins/headers
	go test -v -race -tags=debug ./tests/plugins/logger
	go test -v -race -tags=debug ./tests/plugins/metrics
	go test -v -race -tags=debug ./tests/plugins/redis
	go test -v -race -tags=debug ./tests/plugins/resetter
	go test -v -race -tags=debug ./tests/plugins/rpc
	go test -v -race -tags=debug ./tests/plugins/static
	go test -v -race -tags=debug ./plugins/kv/boltdb
	go test -v -race -tags=debug ./plugins/kv/memory
	go test -v -race -tags=debug ./plugins/kv/memcached
	go test -v -race -tags=debug ./tests/plugins/kv/boltdb
	go test -v -race -tags=debug ./tests/plugins/kv/memory
	go test -v -race -tags=debug ./tests/plugins/kv/memcached
	docker-compose -f tests/docker-compose.yaml down

test_1.14: ## Run application tests
	docker-compose -f tests/docker-compose.yaml up -d
	go1.14.14 test -v -race -tags=debug ./pkg/transport/pipe
	go1.14.14 test -v -race -tags=debug ./pkg/transport/socket
	go1.14.14 test -v -race -tags=debug ./pkg/pool
	go1.14.14 test -v -race -tags=debug ./pkg/worker
	go1.14.14 test -v -race -tags=debug ./pkg/worker_watcher
	go1.14.14 test -v -race -tags=debug ./tests/plugins/http
	go1.14.14 test -v -race -tags=debug ./plugins/http/config
	go1.14.14 test -v -race -tags=debug ./tests/plugins/informer
	go1.14.14 test -v -race -tags=debug ./tests/plugins/reload
	go1.14.14 test -v -race -tags=debug ./tests/plugins/server
	go1.14.14 test -v -race -tags=debug ./tests/plugins/status
	go1.14.14 test -v -race -tags=debug ./tests/plugins/config
	go1.14.14 test -v -race -tags=debug ./tests/plugins/gzip
	go1.14.14 test -v -race -tags=debug ./tests/plugins/headers
	go1.14.14 test -v -race -tags=debug ./tests/plugins/logger
	go1.14.14 test -v -race -tags=debug ./tests/plugins/metrics
	go1.14.14 test -v -race -tags=debug ./tests/plugins/redis
	go1.14.14 test -v -race -tags=debug ./tests/plugins/resetter
	go1.14.14 test -v -race -tags=debug ./tests/plugins/rpc
	go1.14.14 test -v -race -tags=debug ./tests/plugins/static
	go1.14.14 test -v -race -tags=debug ./plugins/kv/boltdb
	go1.14.14 test -v -race -tags=debug ./plugins/kv/memory
	go1.14.14 test -v -race -tags=debug ./plugins/kv/memcached
	go1.14.14 test -v -race -tags=debug ./tests/plugins/kv/boltdb
	go1.14.14 test -v -race -tags=debug ./tests/plugins/kv/memory
	go1.14.14 test -v -race -tags=debug ./tests/plugins/kv/memcached
	docker-compose -f tests/docker-compose.yaml down

test_1.16: ## Run application tests
	docker-compose -f tests/docker-compose.yaml up -d
	go1.16rc1 test -v -race -tags=debug ./pkg/transport/pipe
	go1.16rc1 test -v -race -tags=debug ./pkg/transport/socket
	go1.16rc1 test -v -race -tags=debug ./pkg/pool
	go1.16rc1 test -v -race -tags=debug ./pkg/worker
	go1.16rc1 test -v -race -tags=debug ./pkg/worker_watcher
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/http
	go1.16rc1 test -v -race -tags=debug ./plugins/http/config
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/informer
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/reload
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/server
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/status
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/config
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/gzip
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/headers
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/logger
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/metrics
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/redis
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/resetter
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/rpc
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/static
	go1.16rc1 test -v -race -tags=debug ./plugins/kv/boltdb
	go1.16rc1 test -v -race -tags=debug ./plugins/kv/memory
	go1.16rc1 test -v -race -tags=debug ./plugins/kv/memcached
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/kv/boltdb
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/kv/memory
	go1.16rc1 test -v -race -tags=debug ./tests/plugins/kv/memcached
	docker-compose -f tests/docker-compose.yaml down

test_pipeline: test_1.14 test
