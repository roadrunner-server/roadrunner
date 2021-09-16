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
	docker-compose -f tests/env/docker-compose.yaml down
