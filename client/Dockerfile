FROM node:8.4.0
WORKDIR /home/weave
COPY package.json yarn.lock /home/weave/
ENV NPM_CONFIG_LOGLEVEL=warn NPM_CONFIG_PROGRESS=false
RUN yarn --pure-lockfile
COPY webpack.local.config.js webpack.production.config.js server.js .babelrc .eslintrc .eslintignore .stylelintrc .sass-lint.yml /home/weave/

ARG revision
LABEL maintainer="Weaveworks <help@weave.works>" \
      org.opencontainers.image.title="client" \
      org.opencontainers.image.source="https://github.com/weaveworks/scope" \
      org.opencontainers.image.revision="${revision}" \
      org.opencontainers.image.vendor="Weaveworks"
