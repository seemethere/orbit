PACKAGES=$(shell go list ./... | grep -v /vendor/)
REVISION=$(shell git rev-parse HEAD)

GO_VERBOSE=
ifdef VERBOSE
	GO_VERBOSE=-v
endif

GO          := go
GO_LDFLAGS   =-s -w -X github.com/stellarproject/orbit/version.Version=$(REVISION)
GO_BUILD     =$(GO) build $(GO_VERBOSE) -ldflags '${GO_LDFLAGS}'
BINARIES     =ob orbit-network orbit-log

all: clean $(addprefix bin/,$(BINARIES))
	go build -v -ldflags '${GO_LDFLAGS}' github.com/stellarproject/orbit/agent

bin/%:
	$(GO_BUILD) -o $@ github.com/stellarproject/orbit/cmd/$*

.PHONY: clean
clean:
	$(RM) -r bin/

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
