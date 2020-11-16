test:
	go test -v -race -cover . -tags=debug
	go test -v -race -cover ./plugins/rpc -tags=debug
	go test -v -race -cover ./plugins/rpc/tests -tags=debug
	go test -v -race -cover ./plugins/config/tests -tags=debug
	go test -v -race -cover ./plugins/server/tests -tags=debug
	go test -v -race -cover ./plugins/logger/tests -tags=debug
	go test -v -race -cover ./plugins/metrics/tests -tags=debug