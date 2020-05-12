# Image page: <https://hub.docker.com/_/golang>
FROM golang:1.14.1-alpine3.11

WORKDIR /workspace

ENTRYPOINT["bash", "build.sh"]