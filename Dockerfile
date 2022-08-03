# Image page: <https://hub.docker.com/_/golang>
FROM --platform=${TARGETPLATFORM:-linux/amd64} golang:1.19-alpine as builder

# app version and build date must be passed during image building (version without any prefix).
# e.g.: `docker build --build-arg "APP_VERSION=1.2.3" --build-arg "BUILD_TIME=$(date +%FT%T%z)" .`
ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

COPY . /src

WORKDIR /src

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s \
-X github.com/roadrunner-server/roadrunner/v2/internal/meta.version=$APP_VERSION \
-X github.com/roadrunner-server/roadrunner/v2/internal/meta.buildTime=$BUILD_TIME"

# compile binary file
RUN set -x
RUN go mod download
RUN go mod tidy
RUN CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o ./rr ./cmd/rr
RUN ./rr -v

FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3

RUN apk upgrade --update-cache --available && \
    apk add openssl && \
    rm -rf /var/cache/apk/*

# use same build arguments for image labels
ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="roadrunner"
LABEL org.opencontainers.image.description="High-performance PHP application server, load-balancer, process manager written in Go and powered with plugins"
LABEL org.opencontainers.image.url="https://github.com/roadrunner-server/roadrunner"
LABEL org.opencontainers.image.source="https://github.com/roadrunner-server/roadrunner"
LABEL org.opencontainers.image.vendor="SpiralScout"
LABEL org.opencontainers.image.version="$APP_VERSION"
LABEL org.opencontainers.image.created="$BUILD_TIME"
LABEL org.opencontainers.image.licenses="MIT"

# copy required files from builder image
COPY --from=builder /src/rr /usr/bin/rr
COPY --from=builder /src/.rr.yaml /etc/rr.yaml

# use roadrunner binary as image entrypoint
ENTRYPOINT ["/usr/bin/rr"]
