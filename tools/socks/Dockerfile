FROM gliderlabs/alpine
WORKDIR /
COPY proxy /
EXPOSE 8000
EXPOSE 8080
ENTRYPOINT ["/proxy"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="socks" \
      org.opencontainers.image.source="https://github.com/weaveworks/build-tools/tree/master/socks" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
