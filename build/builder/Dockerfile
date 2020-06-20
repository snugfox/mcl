FROM golang:1.14.4-alpine3.12

VOLUME /go/src/github.com/snugfox/mcl
WORKDIR /go/src/github.com/snugfox/mcl

RUN set -eux; \
    apk add --no-cache build-base ca-certificates clang docker-cli git \
        make wget; \
    go env -w CC=clang CXX=clang++

ARG GORELEASER_VER="0.138.0"
RUN set -eux; \
    wget -q https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh; \
    chmod +x goreleaser.sh; \
    ./goreleaser.sh -b /usr/local/bin v${GORELEASER_VER}; \
    rm goreleaser.sh

COPY go.mod go.sum ./
RUN go mod download
