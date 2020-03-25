#!/usr/bin/env sh

echo "${DOCKERHUB_PASSWORD}" | docker login -u ${DOCKERHUB_USERNAME} --password-stdin
echo "${GITHUB_TOKEN}" | docker login docker.pkg.github.com -u snugfox --password-stdin

goreleaser release --rm-dist
