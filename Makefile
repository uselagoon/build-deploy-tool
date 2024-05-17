VERSION=$(shell echo $(shell git describe --abbrev=0 --tags)+$(shell git rev-parse --short=8 HEAD))
BUILD=$(shell date +%FT%T%z)
GO_VER=1.22
GOOS=linux
GOARCH=amd64
PKG=github.com/uselagoon/build-deploy-tool
LDFLAGS=-w -s -X ${PKG}/cmd.bdtVersion=${VERSION} -X ${PKG}/cmd.bdtBuild=${BUILD} -X "${PKG}/cmd.goVersion=${GO_VER}"

test: fmt vet
	go clean -testcache && go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

run: fmt vet
	go run ./main.go

build: fmt vet
	CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -v

docker-build:
	DOCKER_BUILDKIT=1 docker build --pull --build-arg GO_VER=${GO_VER} --build-arg VERSION=${VERSION} --build-arg BUILD=${BUILD} --rm -f Dockerfile -t build-deploy-image:local .
	docker run --entrypoint /bin/bash build-deploy-image:local -c 'build-deploy-tool version'
