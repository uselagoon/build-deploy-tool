VERSION=$(shell echo $(shell git describe --abbrev=0 --tags))
SHORTCOMMIT=$(shell echo $(shell git rev-parse --short=8 HEAD))
BUILD=$(shell date +%FT%T%z)
GO_VER=$(shell go version)
GOOS=linux
GOARCH=amd64
PKG=github.com/uselagoon/build-deploy-tool
LDFLAGS=-w -s -X ${PKG}/cmd.bdtVersion=${VERSION} -X ${PKG}/cmd.shortCommit=${SHORTCOMMIT} -X ${PKG}/cmd.bdtBuild=${BUILD} -X "${PKG}/cmd.goVersion=${GO_VER}"

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
