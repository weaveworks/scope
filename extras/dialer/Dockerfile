FROM golang:1.7
ADD ./bin/dialer /go/bin

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="dialer" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/dialer" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
