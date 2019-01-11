#!/bin/bash
LDFLAGS="$LDFLAGS -X github.com/spiral/roadrunner/cmd/rr/cmd.Version=${RR_VERSION}"
LDFLAGS="$LDFLAGS -X github.com/spiral/roadrunner/cmd/rr/cmd.BuildTime=$(date +%FT%T%z)"

# Verify all external modules
go mod init

# Build the binary
CGO_ENABLED=0 go build -v -ldflags "$LDFLAGS -extldflags '-static'" -o "rr"