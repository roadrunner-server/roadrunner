# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.15.6-alpine as builder

# app version and build date must be passed during image building (version without any prefix).
# e.g.: `docker build --build-arg "APP_VERSION=1.2.3" --build-arg "BUILD_TIME=$(date +%FT%T%z)" .`
ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s \
-X github.com/spiral/roadrunner/cmd/rr/cmd.Version=$APP_VERSION \
-X github.com/spiral/roadrunner/cmd/rr/cmd.BuildTime=$BUILD_TIME"

COPY . /src

WORKDIR /src

# download dependencies and compile binary file
RUN set -x \
    && go mod download \
    && go mod verify \
    && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o ./rr ./cmd/rr/main.go

# Image page: <https://hub.docker.com/_/alpine>
FROM alpine:3.12

# use same build arguments for image labels
ARG APP_VERSION
ARG BUILD_TIME

LABEL \
    org.opencontainers.image.title="roadrunner" \
    org.opencontainers.image.description="High-performance PHP application server, load-balancer and process manager" \
    org.opencontainers.image.url="https://github.com/spiral/roadrunner" \
    org.opencontainers.image.source="https://github.com/spiral/roadrunner" \
    org.opencontainers.image.vendor="SpiralScout" \
    org.opencontainers.image.version="$APP_VERSION" \
    org.opencontainers.image.created="$BUILD_TIME" \
    org.opencontainers.image.licenses="MIT"

# copy required files from builder image
COPY --from=builder /src/rr /usr/bin/rr
COPY --from=builder /src/.rr.yaml /etc/rr.yaml

# use roadrunner binary as image entrypoint
ENTRYPOINT ["/usr/bin/rr"]
