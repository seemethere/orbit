PACKAGES=$(shell go list ./... | grep -v /vendor/)
REVISION=$(shell git rev-parse HEAD)
GO_LDFLAGS=-s -w -X github.com/stellarproject/orbit/version.Version=$(REVISION)

all:
	rm -f bin/*
	go build -o bin/orbit -v -ldflags '${GO_LDFLAGS}'
	go build -o bin/orbit-network -v -ldflags '${GO_LDFLAGS}' github.com/stellarproject/orbit/orbit-network

vab:
	rm -f bin/*
	vab build --local

static:
	CGO_ENALBED=0 go build -v -ldflags '${GO_LDFLAGS} -extldflags "-static"'

install:
	@install bin/* /usr/local/bin/

FORCE:

plugin: FORCE
	go build -o orbit-linux-amd64.so -v -buildmode=plugin github.com/stellarproject/orbit/plugin/

protos:
	protobuild --quiet ${PACKAGES}
