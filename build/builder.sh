#!/usr/bin/env sh

set -eu

# Parse command and options
cmd="${1-}"
if ! opts=$(getopt -n "$(basename "$0")" -- "$@"); then
	exit 1
fi
eval set -- "$opts"
while true; do
	case "$1" in
		"--")
			shift
			break
			;;
	esac
	shift
done

# Docker image metadata
dockerfile="./build/builder/Dockerfile"
name="snugfox/mcl-builder"
tag="$(cat ./go.sum "$dockerfile" | sha256sum | cut -c-7)"
image="${name}:${tag}"

init() {
	if ! docker image inspect -f "{{ .Id }}" "${image}" > /dev/null 2>&1; then
		if docker pull -q "$image" > /dev/null 2>&1; then
			echo "Pulled ${image}" 1>&2
		else
			docker build -q -f "$dockerfile" -t "$image" . > /dev/null
			echo "Built $image" 1>&2
		fi
	fi
}

case "$cmd" in
	"")
		echo "Must specify a command" 2>&1
		;;
	"image")
		echo "$image"
		;;
	"init")
		init
		;;
	"push")
		init
		docker push "${image}" > /dev/null 2>&1
		echo "Pushed ${image}" 1>&2
		;;
	"run")
		init
		exec docker run --rm \
			-e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD \
			-e GITHUB_TOKEN \
			-v /var/run/docker.sock:/var/run/docker.sock \
			-v "${PWD}:/go/src/github.com/snugfox/mcl" \
			"$image" $@
		;;
	*)
		echo "Unknown command: ${1}" 2>&1
		exit 1
		;;
esac
