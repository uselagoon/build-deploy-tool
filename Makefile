VERSION=$(shell echo $(shell git describe --abbrev=0 --tags)+$(shell git rev-parse --short=8 HEAD))
BUILD=$(shell date +%FT%T%z)
# renovate: datasource=golang-version depName=go
GO_VER=1.26
GOOS=linux
GOARCH=amd64
PKG=github.com/uselagoon/build-deploy-tool
LDFLAGS=-w -s -X ${PKG}/cmd.bdtVersion=${VERSION} -X ${PKG}/cmd.bdtBuild=${BUILD} -X "${PKG}/cmd.goVersion=${GO_VER}"

.PHONY: test
test: fmt vet
	go clean -testcache && go test -v ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: run
run: fmt vet
	go run ./main.go

.PHONY: build
build: fmt vet
	CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -v

.PHONY: docker-build
docker-build:
	DOCKER_BUILDKIT=1 docker build --pull --build-arg GO_VER=${GO_VER} --build-arg VERSION=${VERSION} --build-arg BUILD=${BUILD} --rm -f Dockerfile -t lagoon/build-deploy-image:local .
	docker run --rm --entrypoint /bin/bash lagoon/build-deploy-image:local -c 'build-deploy-tool version'

TRAEFIK_EXT_DIR := third_party/traefik-dynamic-ext

# Pulls the version straight from the traefik/v3 require line in go.mod,
# e.g. "github.com/traefik/traefik/v3 v3.7.8" -> "v3.7.8"
TRAEFIK_VERSION := $(shell awk '/github.com\/traefik\/traefik\/v3 / {print $$2}' go.mod)

.PHONY: update-traefik-ext
update-traefik-ext:
ifeq ($(TRAEFIK_VERSION),)
	$(error Could not find github.com/traefik/traefik/v3 in go.mod)
endif
	@echo "Vendoring traefik/dynamic/ext @ $(TRAEFIK_VERSION)"
	@mkdir -p $(TRAEFIK_EXT_DIR)
	@curl -sSLf https://raw.githubusercontent.com/traefik/traefik/$(TRAEFIK_VERSION)/pkg/config/dynamic/ext/ext.go \
		-o $(TRAEFIK_EXT_DIR)/ext.go
	@curl -sSLf https://raw.githubusercontent.com/traefik/traefik/$(TRAEFIK_VERSION)/pkg/config/dynamic/ext/go.mod \
		-o $(TRAEFIK_EXT_DIR)/go.mod
	@go mod tidy