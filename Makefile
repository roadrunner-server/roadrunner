all:
	@./build.sh
	@./build-ce.sh
build:
	@./build.sh all
build-ce:
	@./build-ce.sh all
clean:
	rm -rf rr
	rm -rf rr-ce
install: all
	cp rr /usr/local/bin/rr
	cp rr-ce /usr/local/bin/rr-ce
uninstall: 
	rm -f /usr/local/bin/rr
	rm -f /usr/local/bin/rr-ce
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
