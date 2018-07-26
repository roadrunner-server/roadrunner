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
	go test -v -race -cover
	go test -v -race -cover ./service
	go test -v -race -cover ./service/env
	go test -v -race -cover ./service/rpc
	go test -v -race -cover ./service/http
	go test -v -race -cover ./service/static
