FROM golang:1.12 as containerd

RUN apt-get update && \
	apt-get install -y \
	rsync \
	autoconf \
	automake \
	g++ \
	libtool \
	unzip \
	btrfs-tools \
	gcc \
	git \
	libseccomp-dev \
	make \
	xfsprogs

ENV CONTAINERD_VERSION 2d8e60b8f22deab9c1f161dac427b11c37a7f936
RUN git clone https://github.com/crosbymichael/containerd /go/src/github.com/containerd/containerd
RUN git clone https://github.com/opencontainers/runc /go/src/github.com/opencontainers/runc

WORKDIR /go/src/github.com/containerd/containerd
RUN git checkout ${CONTAINERD_VERSION}

RUN rsync -au /go/src/github.com/containerd/containerd/vendor/ /go/src/ && \
	rm -rf /go/src/github.com/containerd/containerd/vendor/

RUN ./script/setup/install-protobuf
RUN ./script/setup/install-runc
RUN make

FROM containerd as orbit

ADD . /go/src/github.com/stellarproject/orbit
RUN	rm -rf /go/src/github.com/stellarproject/orbit/vendor/github.com/containerd/ && \
	rsync -au --ignore-existing /go/src/github.com/stellarproject/orbit/vendor/ /go/src/ && \
	rm -rf /go/src/github.com/stellarproject/orbit/vendor/

WORKDIR /go/src/github.com/stellarproject/orbit

RUN make && make plugin

FROM scratch

COPY --from=containerd /go/src/github.com/containerd/containerd/bin/* /bin/
COPY --from=containerd /usr/local/sbin/runc /sbin/
COPY --from=orbit /go/src/github.com/stellarproject/orbit/bin/* /bin/
COPY --from=orbit /go/src/github.com/stellarproject/orbit/plugins/* /plugins/
