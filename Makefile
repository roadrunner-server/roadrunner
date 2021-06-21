#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test_coverage:
	docker-compose -f tests/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage
	mkdir coverage
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/pipe.out -covermode=atomic ./pkg/transport/pipe
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/socket.out -covermode=atomic ./pkg/transport/socket
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/pool.out -covermode=atomic ./pkg/pool
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/worker.out -covermode=atomic ./pkg/worker
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/worker_stack.out -covermode=atomic ./pkg/worker_watcher
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/bst.out -covermode=atomic ./pkg/bst
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/http.out -covermode=atomic ./tests/plugins/http
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/http_config.out -covermode=atomic ./plugins/http/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/server_cmd.out -covermode=atomic ./plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/informer.out -covermode=atomic ./tests/plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/reload.out -covermode=atomic ./tests/plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/server.out -covermode=atomic ./tests/plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/service.out -covermode=atomic ./tests/plugins/service
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/status.out -covermode=atomic ./tests/plugins/status
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/config.out -covermode=atomic ./tests/plugins/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/gzip.out -covermode=atomic ./tests/plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/headers.out -covermode=atomic ./tests/plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/logger.out -covermode=atomic ./tests/plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/metrics.out -covermode=atomic ./tests/plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/redis.out -covermode=atomic ./tests/plugins/redis
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/resetter.out -covermode=atomic ./tests/plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/rpc.out -covermode=atomic ./tests/plugins/rpc
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/kv_plugin.out -covermode=atomic ./tests/plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/broadcast_plugin.out -covermode=atomic ./tests/plugins/broadcast
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/ws_plugin.out -covermode=atomic ./tests/plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage/ws_origin.out -covermode=atomic ./plugins/websockets
	cat ./coverage/*.out > ./coverage/summary.out
	docker-compose -f tests/docker-compose.yaml down

test: ## Run application tests
	docker-compose -f tests/docker-compose.yaml up -d
	go test -v -race -tags=debug ./pkg/transport/pipe
	go test -v -race -tags=debug ./pkg/transport/socket
	go test -v -race -tags=debug ./pkg/pool
	go test -v -race -tags=debug ./pkg/worker
	go test -v -race -tags=debug ./pkg/worker_watcher
	go test -v -race -tags=debug ./pkg/bst
	go test -v -race -tags=debug ./tests/plugins/http
	go test -v -race -tags=debug ./plugins/http/config
	go test -v -race -tags=debug ./plugins/server
	go test -v -race -tags=debug ./tests/plugins/informer
	go test -v -race -tags=debug ./tests/plugins/reload
	go test -v -race -tags=debug ./tests/plugins/server
	go test -v -race -tags=debug ./tests/plugins/service
	go test -v -race -tags=debug ./tests/plugins/status
	go test -v -race -tags=debug ./tests/plugins/config
	go test -v -race -tags=debug ./tests/plugins/gzip
	go test -v -race -tags=debug ./tests/plugins/headers
	go test -v -race -tags=debug ./tests/plugins/logger
	go test -v -race -tags=debug ./tests/plugins/metrics
	go test -v -race -tags=debug ./tests/plugins/redis
	go test -v -race -tags=debug ./tests/plugins/resetter
	go test -v -race -tags=debug ./tests/plugins/rpc
	go test -v -race -tags=debug ./tests/plugins/kv
	go test -v -race -tags=debug ./tests/plugins/broadcast
	go test -v -race -tags=debug ./tests/plugins/websockets
	go test -v -race -tags=debug ./plugins/websockets
	docker-compose -f tests/docker-compose.yaml down

testGo1.17beta1: ## Run application tests
	docker-compose -f tests/docker-compose.yaml up -d
	go1.17beta1 test -v -race -tags=debug ./pkg/transport/pipe
	go1.17beta1 test -v -race -tags=debug ./pkg/transport/socket
	go1.17beta1 test -v -race -tags=debug ./pkg/pool
	go1.17beta1 test -v -race -tags=debug ./pkg/worker
	go1.17beta1 test -v -race -tags=debug ./pkg/worker_watcher
	go1.17beta1 test -v -race -tags=debug ./pkg/bst
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/http
	go1.17beta1 test -v -race -tags=debug ./plugins/http/config
	go1.17beta1 test -v -race -tags=debug ./plugins/server
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/informer
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/reload
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/server
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/service
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/status
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/config
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/gzip
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/headers
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/logger
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/metrics
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/redis
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/resetter
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/rpc
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/kv
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/websockets
	go1.17beta1 test -v -race -tags=debug ./tests/plugins/broadcast
	go1.17beta1 test -v -race -tags=debug ./plugins/websockets
	docker-compose -f tests/docker-compose.yaml down
