FROM golang:1.14.3 as builder

COPY . /src

WORKDIR /src

RUN set -x \
    && apt-get update -y \
    && apt-get install -y bash git \
    && go version \
    && bash ./build.sh \
    && test -f ./.rr.yaml

FROM alpine:latest

LABEL \
    org.opencontainers.image.title="roadrunner" \
    org.opencontainers.image.description="High-performance PHP application server, load-balancer and process manager" \
    org.opencontainers.image.url="https://github.com/spiral/roadrunner" \
    org.opencontainers.image.source="https://github.com/spiral/roadrunner" \
    org.opencontainers.image.vendor="SpiralScout" \
    org.opencontainers.image.licenses="MIT"

COPY --from=builder /src/rr /usr/bin/rr
COPY --from=builder /src/.rr.yaml /etc/rr.yaml

ENTRYPOINT ["/usr/bin/rr"]
