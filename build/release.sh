#!/usr/bin/env sh

BUILD_IMAGE="$1"

echo "${DOCKERHUB_PASSWORD}" | docker login -u ${DOCKERHUB_USERNAME} --password-stdin
echo "${GITHUB_TOKEN}" | docker login docker.pkg.github.com -u snugfox --password-stdin

docker run --rm -e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD -e GITHUB_TOKEN \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "$(pwd):/go/src/github.com/snugfox/mcl" ${BUILD_IMAGE} \
    goreleaser release --rm-dist
