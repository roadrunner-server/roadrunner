# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.13-alpine as builder

COPY . /src

WORKDIR /src

RUN set -x \
    && apk add --no-cache bash git \
    && go version \
    && bash ./build.sh \
    && test -f ./.rr.yaml

FROM alpine:latest

LABEL \
    org.label-schema.name="roadrunner" \
    org.label-schema.description="High-performance PHP application server, load-balancer and process manager" \
    org.label-schema.url="https://github.com/spiral/roadrunner" \
    org.label-schema.vcs-url="https://github.com/spiral/roadrunner" \
    org.label-schema.vendor="SpiralScout" \
    org.label-schema.license="MIT" \
    org.label-schema.schema-version="1.0"

COPY --from=builder /src/rr /usr/bin/rr
COPY --from=builder /src/.rr.yaml /etc/rr.yaml

ENTRYPOINT ["/usr/bin/rr"]
