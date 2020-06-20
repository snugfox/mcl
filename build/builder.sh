#!/usr/bin/env sh

set -eu

dockerfile="./build/builder/Dockerfile"
name="docker.pkg.github.com/snugfox/mcl/builder"
tag="$(cat ./go.sum "$dockerfile" | sha256sum | cut -c-7)"
image="${name}:${tag}"

case "${1-}" in
	"image")
		echo "$image"
		exit 0
		;;
esac

if ! docker image inspect -f "{{ .Id }}" "${image}" > /dev/null 2>&1; then
	if docker pull -q "$image" > /dev/null 2>&1; then
		echo "Pulled ${image}" 1>&2
	else
		docker build -q -f "$dockerfile" -t "$image" . > /dev/null
		echo "Built $image" 1>&2
	fi
fi

case "${1-}" in
	"push")
		docker push "${image}" > /dev/null 2>&1
		echo "Pushed ${image}"
		;;
	"run")
		shift
		exec docker run --rm \
			-e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD \
			-e GITHUB_TOKEN \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v "${PWD}:/go/src/github.com/snugfox/mcl" \
			"$image" $@
		;;
esac
