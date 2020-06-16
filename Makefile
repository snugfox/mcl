DOCKER     := docker
GO         := go
GORELEASER := goreleaser
SHELL      := /bin/sh # Use POSIX shell for portability

# Go version metadata required for build
export GOVER := $(shell $(GO) version | cut -d" " -f3)


.PHONY: build
build:
	$(GORELEASER) build --rm-dist


.PHONY: build-snapshot
build-snapshot:
	$(GORELEASER) build --rm-dist --snapshot


.PHONY: release
release: test
		echo "$$DOCKERHUB_PASSWORD" | $(DOCKER) login -u "$$DOCKERHUB_USERNAME" --password-stdin
		echo "$$GITHUB_TOKEN" | $(DOCKER) login docker.pkg.github.com -u snugfox --password-stdin
		$(GORELEASER) release --rm-dist


.PHONY: release-snapshot
release-snapshot: test
	$(GORELEASER) release --rm-dist --snapshot


.PHONY: test
test:
	$(GO) test -race ./...
