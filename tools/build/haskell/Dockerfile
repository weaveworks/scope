FROM fpco/stack-build:lts-8.9
COPY build.sh /
COPY copy-libraries /usr/local/bin/
ENTRYPOINT ["/build.sh"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="haskell" \
      org.opencontainers.image.source="https://github.com/weaveworks/build-tools/tree/master/build/haskell" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
