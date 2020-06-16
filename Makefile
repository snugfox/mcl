DOCKER     := docker
GO         := go
GORELEASER := goreleaser
SHELL      := /bin/sh # Use POSIX shell for portability

# Build environment, which supports either local or docker.
BUILD_ENV     := local
build_env_dep  = build-env-$(BUILD_ENV)
build_env_cmd  =

ifeq "$(BUILD_ENV)" "docker"
	build_env_cmd  = docker run $(DOCKER_FLAGS) mcl-build
endif

# Docker build environment
DOCKER_FLAGS = --rm -it \
							 -e DOCKERHUB_USERNAME -e DOCKERHUB_PASSWORD \
							 -e GITHUB_TOKEN \
               -v /var/run/docker.sock:/var/run/docker.sock \
               -v $$PWD:/go/src/github.com/snugfox/mcl

# Go version metadata required for build
export GOVER := $(shell $(GO) version | cut -d" " -f3)
DOCKER_FLAGS += -e GOVER


.PHONY: build-env-local
build-env-local:


.PHONY: build-env-docker
build-env-docker:
	$(DOCKER) build -f ./build/build.Dockerfile -t mcl-build .


.PHONY: build
build: $(build_env_dep)
	$(build_env_cmd) $(GORELEASER) build --rm-dist


.PHONY: build-snapshot
build-snapshot: $(build_env_dep)
	$(build_env_cmd) $(GORELEASER) build --rm-dist --snapshot


.PHONY: release
release: test $(build_env_dep)
	$(build_env_cmd) sh -c ' \
		set -eu; \
		echo "$$DOCKERHUB_PASSWORD" | docker login -u "$$DOCKERHUB_USERNAME" --password-stdin; \
		echo "$$GITHUB_TOKEN" | docker login docker.pkg.github.com -u snugfox --password-stdin; \
		$(GORELEASER) release --rm-dist'


.PHONY: release-snapshot
release-snapshot: test $(build_env_dep)
	$(build_env_cmd) $(GORELEASER) release --rm-dist --snapshot


.PHONY: test
test:
	$(GO) test ./...
