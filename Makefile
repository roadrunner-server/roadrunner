all:
	@./build.sh
build:
	@./build.sh all
clean:
	rm -rf rr
install: all
	cp rr /usr/local/bin/rr
uninstall:
	rm -f /usr/local/bin/rr
test:
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
