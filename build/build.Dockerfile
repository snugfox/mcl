FROM golang:1.14.1-alpine3.11

RUN set -eux; \
    apk --no-cache upgrade; \
    apk add --no-cache git docker ca-certificates tzdata

ARG GORELEASER_VER="0.129.0"
RUN set -eux; \
    wget -q https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh; \
    chmod +x goreleaser.sh; \
    ./goreleaser.sh -b /usr/local/bin v${GORELEASER_VER}; \
    rm goreleaser.sh

WORKDIR /go/src/github.com/snugfox/mcl
