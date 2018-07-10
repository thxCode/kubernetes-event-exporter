#############
# phase one #
#############
FROM golang:1.10.3-alpine3.7 AS builder

RUN apk add --no-cache --update \
	    curl \
        git \
    ; \
    curl -k https://glide.sh/get | sh; \
    chmod +x /usr/local/bin/dep; \
    go get -u github.com/prometheus/promu; \
    git clone https://github.com/thxcode/kubernetes-event-exporter.git $GOPATH/src/github.com/thxcode/kubernetes-event-exporter

## build
RUN cd $GOPATH/src/github.com/thxcode/kubernetes-event-exporter; \
    glide install --skip-test; \
    $GOPATH/bin/promu build --prefix ./bin; \
    mkdir -p /build; \
    cp -f ./bin/kubernetes-event-exporter /build/

#############
# phase two #
#############
FROM alpine:3.7

MAINTAINER Frank Mai <frank@rancher.com>

ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

LABEL \
    io.github.thxcode.build-date=$BUILD_DATE \
    io.github.thxcode.name="kubernetes-event-exporter" \
    io.github.thxcode.description="An exporter exposes events of Kubernetes." \
    io.github.thxcode.url="https://github.com/thxcode/kubernetes-event-exporter" \
    io.github.thxcode.vcs-type="Git" \
    io.github.thxcode.vcs-ref=$VCS_REF \
    io.github.thxcode.vcs-url="https://github.com/thxcode/kubernetes-event-exporter.git" \
    io.github.thxcode.vendor="Rancher Labs, Inc" \
    io.github.thxcode.version=$VERSION \
    io.github.thxcode.schema-version="1.0" \
    io.github.thxcode.license="MIT" \
    io.github.thxcode.docker.dockerfile="/Dockerfile"

RUN apk add --no-cache --update \
        ca-certificates \
    ; \
    mkdir -p /data; \
    chown -R nobody:nogroup /data

COPY --from=builder /build/kubernetes-event-exporter /usr/sbin/kubernetes-event-exporter

USER    nobody
EXPOSE  9173 80
VOLUME  [ "/data" ]

ENTRYPOINT [ "/usr/sbin/kubernetes-event-exporter" ]
