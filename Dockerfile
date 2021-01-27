# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.15.7 as builder

# app version and build date must be passed during image building (version without any prefix).
# e.g.: `docker build --build-arg "APP_VERSION=1.2.3" --build-arg "BUILD_TIME=$(date +%FT%T%z)" .`
ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s \
-X github.com/spiral/roadrunner/v2/cmd/cli.Version=$APP_VERSION \
-X github.com/spiral/roadrunner/v2/cmd/cli.BuildTime=$BUILD_TIME"

RUN mkdir /src

WORKDIR /src
RUN mkdir -p /src/cli

COPY cmd/go.mod ./
COPY cmd/go.sum ./
COPY cmd/main.go ./
COPY cmd/cli/* /src/cli/
COPY .rr.yaml ./

# Burn modules cache
RUN set -x \
    && go version \
    && go mod download \
    && go mod verify

# compile binary file
RUN CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o rr main.go

# Image page: <https://hub.docker.com/_/alpine>
FROM alpine:3.13

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
