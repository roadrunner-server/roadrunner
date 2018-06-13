test:
	go test -v -race -cover
	go test -v -race -cover ./service
	go test -v -race -cover ./service/rpc
	go test -v -race -cover ./service/http
	go test -v -race -cover ./service/static