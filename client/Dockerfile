# Changes to this file will not take effect in CI
# until the image version in the CI config is updated. See
# https://github.com/weaveworks/scope/blob/master/.circleci/config.yml#L11
FROM node:8.11
ENV NPM_CONFIG_LOGLEVEL=warn
ENV NPM_CONFIG_PROGRESS=false
ENV XDG_CACHE_HOME=/home/weave/scope/.cache
WORKDIR /home/weave/scope/client

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="scope-ui-build" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
