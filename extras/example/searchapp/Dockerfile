FROM progrium/busybox
WORKDIR /home/weave
ADD searchapp /home/weave/
EXPOSE 8080
ENTRYPOINT ["/home/weave/searchapp"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="example-searchapp" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/example/searchapp" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
