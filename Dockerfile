# Image page: <https://hub.docker.com/_/golang>
FROM --platform=${TARGETPLATFORM:-linux/amd64} golang:1.26-alpine AS builder

# app version and build date must be passed during image building (version without any prefix).
# e.g.: `docker build --build-arg "APP_VERSION=1.2.3" --build-arg "BUILD_TIME=$(date +%FT%T%z)" .`
ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

WORKDIR /src

# Copy module files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .

# arguments to pass on each go tool link invocation
ENV LDFLAGS="-s \
	-X github.com/roadrunner-server/roadrunner/v2025/internal/meta.version=$APP_VERSION \
	-X github.com/roadrunner-server/roadrunner/v2025/internal/meta.buildTime=$BUILD_TIME"

# compile and verify binary
RUN CGO_ENABLED=0 go build -trimpath -ldflags "$LDFLAGS" -o ./rr ./cmd/rr && ./rr -v

# ---- Final stage ----
FROM --platform=${TARGETPLATFORM:-linux/amd64} alpine:3

RUN apk add --no-cache ca-certificates

# use same build arguments for image labels
ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

# https://github.com/opencontainers/image-spec/blob/main/annotations.md
LABEL org.opencontainers.image.title="roadrunner"
LABEL org.opencontainers.image.description="High-performance PHP application server and process manager written in Go and powered with plugins"
LABEL org.opencontainers.image.url="https://roadrunner.dev"
LABEL org.opencontainers.image.source="https://github.com/roadrunner-server/roadrunner"
LABEL org.opencontainers.image.vendor="SpiralScout"
LABEL org.opencontainers.image.version="$APP_VERSION"
LABEL org.opencontainers.image.created="$BUILD_TIME"
LABEL org.opencontainers.image.licenses="MIT"

# Non-root user
RUN addgroup -S rr && adduser -S -G rr rr

# copy required files from builder image
COPY --from=builder /src/rr /usr/bin/rr
COPY --from=builder /src/.rr.yaml /etc/rr.yaml

USER rr

# use roadrunner binary as image entrypoint
ENTRYPOINT ["/usr/bin/rr"]
CMD ["serve", "-c", "/etc/rr.yaml"]
