FROM golang:1.10.1-stretch

RUN apt-get update && apt-get install jq pv \
  && chmod a+wr --recursive /usr/local/go/pkg \
  && (curl https://sdk.cloud.google.com | bash) \
  && pip install --upgrade requests awscli


ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="build-image" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope/tree/master/build-image" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
