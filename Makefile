test:
	go test -v -race ./...

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s" -o rr cmd/rr/main.go

debug:
	dlv debug cmd/rr/main.go -- serve -c .rr-sample-bench-http.yaml
