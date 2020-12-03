test:
	go clean -testcache
	go test -v -race -cover . -tags=debug
	go test -v -race -cover ./plugins/rpc -tags=debug
	go test -v -race -cover ./plugins/rpc/tests -tags=debug
	go test -v -race -cover ./plugins/config/tests -tags=debug
	go test -v -race -cover ./plugins/server/tests -tags=debug
	go test -v -race -cover ./plugins/logger/tests -tags=debug
	go test -v -race -cover ./plugins/metrics/tests -tags=debug
	go test -v -race -cover ./plugins/informer/tests -tags=debug
	go test -v -race -cover ./plugins/resetter/tests -tags=debug
	go test -v -race -cover ./plugins/http/attributes -tags=debug
	go test -v -race -cover ./plugins/http/tests -tags=debug
	go test -v -race -cover ./plugins/gzip/tests -tags=debug
	go test -v -race -cover ./plugins/static/tests -tags=debug
	go test -v -race -cover ./plugins/static -tags=debug
	go test -v -race -cover ./plugins/headers/tests -tags=debug
	go test -v -race -cover ./plugins/checker/tests -tags=debug

test_headers:
	go test -v -race -cover ./plugins/headers/tests -tags=debug
test_checker:
	go test -v -race -cover ./plugins/checker/tests -tags=debug