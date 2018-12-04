FROM ubuntu
WORKDIR /home/weave
ADD ./qotd /home/weave/
EXPOSE 4446
ENTRYPOINT ["/home/weave/qotd"]

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="example-qotd" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/extras/example/qotd" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
