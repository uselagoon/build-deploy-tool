VERSION=$(shell echo $(shell git describe --abbrev=0 --tags)+$(shell git rev-parse --short=8 HEAD))
BUILD=$(shell date +%FT%T%z)
GOCMD=go
GO_VER=1.22
GOOS=linux
GOARCH=amd64
PKG=github.com/uselagoon/build-deploy-tool
LDFLAGS=-w -s -X ${PKG}/cmd.bdtVersion=${VERSION} -X ${PKG}/cmd.bdtBuild=${BUILD} -X "${PKG}/cmd.goVersion=${GO_VER}"

test: fmt vet
	$(GOCMD) clean -testcache && $(GOCMD) test -v ./...

fmt:
	$(GOCMD) fmt ./...

vet:
	$(GOCMD) vet ./...

run: fmt vet
	$(GOCMD) run ./main.go

build: fmt vet
	CGO_ENABLED=0 $(GOCMD) build -ldflags '${LDFLAGS}' -v

docker-build:
	DOCKER_BUILDKIT=1 docker build --pull --build-arg GO_VER=${GO_VER} --build-arg VERSION=${VERSION} --build-arg BUILD=${BUILD} --rm -f Dockerfile -t build-deploy-image:local .
	docker run --entrypoint /bin/bash build-deploy-image:local -c 'build-deploy-tool version'

docs: test
	$(GOCMD) run main.go --docs

MKDOCS_IMAGE ?= ghcr.io/amazeeio/mkdocs-material
MKDOCS_SERVE_PORT ?= 8000

.PHONY: docs/serve
docs/serve:
	@echo "Starting container to serve documentation"
	@docker run --rm -it \
		-p 127.0.0.1:$(MKDOCS_SERVE_PORT):$(MKDOCS_SERVE_PORT) \
		-v ${PWD}:/docs \
		--entrypoint sh $(MKDOCS_IMAGE) \
		-c 'mkdocs serve --dev-addr=0.0.0.0:$(MKDOCS_SERVE_PORT) -f mkdocs.yml'