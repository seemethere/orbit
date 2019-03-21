PACKAGES=$(shell go list ./... | grep -v /vendor/)
REVISION=$(shell git rev-parse HEAD)
GO_LDFLAGS=-s -w -X github.com/stellarproject/orbit/version.Version=$(REVISION)

all:
	rm -f bin/*
	go build -v -ldflags '${GO_LDFLAGS}' github.com/stellarproject/orbit/agent
	go build -o bin/ob -v -ldflags '${GO_LDFLAGS}' github.com/stellarproject/orbit/cmd/ob
	go build -o bin/orbit-network -v -ldflags '${GO_LDFLAGS}' github.com/stellarproject/orbit/cmd/orbit-network
	go build -o bin/orbit-log -v -ldflags '${GO_LDFLAGS}' github.com/stellarproject/orbit/cmd/orbit-log

vab:
	rm -f bin/*
	vab build --local

static:
	CGO_ENALBED=0 go build -v -ldflags '${GO_LDFLAGS} -extldflags "-static"'

install:
	@install bin/* /usr/local/bin/
	@install plugins/* /var/lib/containerd/plugins/

FORCE:

plugin: FORCE
	go build -o plugins/orbit-linux-amd64.so -ldflags '${GO_LDFLAGS}' -v -buildmode=plugin github.com/stellarproject/orbit/plugin/

protos:
	protobuild --quiet ${PACKAGES}
