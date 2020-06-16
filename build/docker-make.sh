#!/usr/bin/env sh

set -eu

echo "Building Docker build image..." >&2
image_id=$(docker build -q -f ./build/Dockerfile.build .)
echo "Built image $image_id" >&2

echo "Running make in build image..." >&2
exec docker run --rm -it \
    -e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD \
    -e GITHUB_TOKEN \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v "${PWD}:/go/src/github.com/snugfox/mcl" \
    $image_id make $@
