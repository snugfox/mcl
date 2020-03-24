#!/usr/bin/env sh

BUILD_IMAGE="$1"

docker run --rm -e DOCKER_USERNAME -e DOCKER_PASSWORD -e GITHUB_TOKEN \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$(pwd):/go/src/github.com/snugfox/mcl" ${BUILD_IMAGE} \
    goreleaser release --rm-dist
