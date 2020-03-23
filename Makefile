build:
	@./build.sh
all:
	@./build.sh all
clean:
	rm -rf rr
install: all
	cp rr /usr/local/bin/rr
uninstall:
	rm -f /usr/local/bin/rr
test:
	go mod vendor
	composer update
	go test -v -race -cover
	go test -v -race -cover ./util
	go test -v -race -cover ./service
	go test -v -race -cover ./service/env
	go test -v -race -cover ./service/rpc
	go test -v -race -cover ./service/http
	go test -v -race -cover ./service/static
	go test -v -race -cover ./service/limit
	go test -v -race -cover ./service/headers
	go test -v -race -cover ./service/metrics
	go test -v -race -cover ./service/health
	go test -v -race -cover ./service/compression
	go test -v -race -cover ./service/reload
lint:
	go fmt ./...
	golint ./...