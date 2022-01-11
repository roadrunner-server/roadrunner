#!/usr/bin/make
# Makefile readme (ru): <http://linux.yaroslavl.ru/docs/prog/gnu_make_3-79_russian_manual.html>
# Makefile readme (en): <https://www.gnu.org/software/make/manual/html_node/index.html#SEC_Contents>

SHELL = /bin/sh

test_coverage:
	rm -rf coverage-ci
	mkdir ./coverage-ci
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pipe.out -covermode=atomic ./ipc/pipe
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/socket.out -covermode=atomic ./ipc/socket
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pool.out -covermode=atomic ./pool
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/worker.out -covermode=atomic ./worker
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/bst.out -covermode=atomic ./bst
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/pq.out -covermode=atomic ./priority_queue
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/worker_stack.out -covermode=atomic ./worker_watcher
	go test -v -race -cover -tags=debug -coverpkg=./... -coverprofile=./coverage-ci/events.out -covermode=atomic ./events
	echo 'mode: atomic' > ./coverage-ci/summary.txt
	tail -q -n +2 ./coverage-ci/*.out >> ./coverage-ci/summary.txt

test: ## Run application tests
	go test -v -race -tags=debug ./ipc/pipe
	go test -v -race -tags=debug ./ipc/socket
	go test -v -race -tags=debug ./pool
	go test -v -race -tags=debug ./worker
	go test -v -race -tags=debug ./worker_watcher
	go test -v -race -tags=debug ./bst
	go test -v -race -tags=debug ./priority_queue
	go test -v -race -tags=debug ./events
