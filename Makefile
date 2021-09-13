#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test_coverage:
	docker-compose -f tests/env/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage-ci
	mkdir ./coverage-ci
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pipe.out -covermode=atomic ./pkg/transport/pipe
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/socket.out -covermode=atomic ./pkg/transport/socket
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pool.out -covermode=atomic ./pkg/pool
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/worker.out -covermode=atomic ./pkg/worker
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/bst.out -covermode=atomic ./pkg/bst
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pq.out -covermode=atomic ./pkg/priority_queue
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/worker_stack.out -covermode=atomic ./pkg/worker_watcher
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/ws_origin.out -covermode=atomic ./plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/http_config.out -covermode=atomic ./plugins/http/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/server_cmd.out -covermode=atomic ./plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/struct_jobs.out -covermode=atomic ./plugins/jobs/job
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pipeline_jobs.out -covermode=atomic ./plugins/jobs/pipeline
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/jobs_core.out -covermode=atomic ./tests/plugins/jobs
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/kv_plugin.out -covermode=atomic ./tests/plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/broadcast_plugin.out -covermode=atomic ./tests/plugins/broadcast
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/websockets.out -covermode=atomic ./tests/plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/http.out -covermode=atomic ./tests/plugins/http
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/informer.out -covermode=atomic ./tests/plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/reload.out -covermode=atomic ./tests/plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/server.out -covermode=atomic ./tests/plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/service.out -covermode=atomic ./tests/plugins/service
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/status.out -covermode=atomic ./tests/plugins/status
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/config.out -covermode=atomic ./tests/plugins/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/gzip.out -covermode=atomic ./tests/plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/headers.out -covermode=atomic ./tests/plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/logger.out -covermode=atomic ./tests/plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/metrics.out -covermode=atomic ./tests/plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/resetter.out -covermode=atomic ./tests/plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/rpc.out -covermode=atomic ./tests/plugins/rpc
	echo 'mode: atomic' > ./coverage-ci/summary.txt
	tail -q -n +2 ./coverage-ci/*.out >> ./coverage-ci/summary.txt
	docker-compose -f tests/env/docker-compose.yaml down

test: ## Run application tests
	docker-compose -f tests/env/docker-compose.yaml up -d
	go test -v -race -tags=debug ./pkg/transport/pipe
	go test -v -race -tags=debug ./pkg/transport/socket
	go test -v -race -tags=debug ./pkg/pool
	go test -v -race -tags=debug ./pkg/worker
	go test -v -race -tags=debug ./pkg/worker_watcher
	go test -v -race -tags=debug ./pkg/bst
	go test -v -race -tags=debug ./pkg/priority_queue
	go test -v -race -tags=debug ./plugins/jobs/pipeline
	go test -v -race -tags=debug ./plugins/http/config
	go test -v -race -tags=debug ./plugins/server
	go test -v -race -tags=debug ./plugins/jobs/job
	go test -v -race -tags=debug ./tests/plugins/jobs
	go test -v -race -tags=debug ./tests/plugins/kv
	go test -v -race -tags=debug ./tests/plugins/broadcast
	go test -v -race -tags=debug ./tests/plugins/websockets
	go test -v -race -tags=debug ./plugins/websockets
	go test -v -race -tags=debug ./tests/plugins/http
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
	go test -v -race -tags=debug ./tests/plugins/resetter
	go test -v -race -tags=debug ./tests/plugins/rpc
	docker-compose -f tests/env/docker-compose.yaml down
