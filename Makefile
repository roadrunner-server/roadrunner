#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test_coverage:
	docker-compose -f tests/env/docker-compose.yaml up -d --remove-orphans
	rm -rf coverage-ci
	mkdir ./coverage-ci
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pipe.txt -covermode=atomic ./pkg/transport/pipe
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/socket.txt -covermode=atomic ./pkg/transport/socket
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pool.txt -covermode=atomic ./pkg/pool
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/worker.txt -covermode=atomic ./pkg/worker
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/bst.txt -covermode=atomic ./pkg/bst
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pq.txt -covermode=atomic ./pkg/priority_queue
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/worker_stack.txt -covermode=atomic ./pkg/worker_watcher
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/ws_origin.txt -covermode=atomic ./plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/http_config.txt -covermode=atomic ./plugins/http/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/server_cmd.txt -covermode=atomic ./plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/struct_jobs.txt -covermode=atomic ./plugins/jobs/job
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pipeline_jobs.txt -covermode=atomic ./plugins/jobs/pipeline
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/jobs_core.txt -covermode=atomic ./tests/plugins/jobs
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/kv_plugin.txt -covermode=atomic ./tests/plugins/kv
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/broadcast_plugin.txt -covermode=atomic ./tests/plugins/broadcast
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/websockets.txt -covermode=atomic ./tests/plugins/websockets
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/http.txt -covermode=atomic ./tests/plugins/http
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/informer.txt -covermode=atomic ./tests/plugins/informer
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/reload.txt -covermode=atomic ./tests/plugins/reload
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/server.txt -covermode=atomic ./tests/plugins/server
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/service.txt -covermode=atomic ./tests/plugins/service
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/status.txt -covermode=atomic ./tests/plugins/status
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/config.txt -covermode=atomic ./tests/plugins/config
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/gzip.txt -covermode=atomic ./tests/plugins/gzip
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/headers.txt -covermode=atomic ./tests/plugins/headers
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/logger.txt -covermode=atomic ./tests/plugins/logger
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/metrics.txt -covermode=atomic ./tests/plugins/metrics
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/resetter.txt -covermode=atomic ./tests/plugins/resetter
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/rpc.txt -covermode=atomic ./tests/plugins/rpc
	cat ./coverage-ci/*.txt > ./coverage-ci/summary.txt
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
