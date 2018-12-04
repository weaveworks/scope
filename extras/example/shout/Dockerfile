FROM alpine:latest
WORKDIR /home/weave
ADD shout /home/weave/
EXPOSE 8090
ENTRYPOINT ["/home/weave/shout"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="example-shout" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/example/shout" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
