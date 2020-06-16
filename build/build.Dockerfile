FROM golang:1.14.4-alpine3.12

VOLUME /go/src/github.com/snugfox/mcl
WORKDIR /go/src/github.com/snugfox/mcl

ENV GOCACHE=/tmp/go-build

RUN apk add --no-cache ca-certificates docker-cli git wget

ARG GORELEASER_VER="0.138.0"
RUN set -eux; \
    wget -q https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh; \
    chmod +x goreleaser.sh; \
    ./goreleaser.sh -b /usr/local/bin v${GORELEASER_VER}; \
    rm goreleaser.sh

COPY go.mod go.sum ./
RUN go mod download
