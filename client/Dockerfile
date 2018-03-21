FROM node:8.4.0
WORKDIR /home/weave
COPY package.json yarn.lock /home/weave/
ENV NPM_CONFIG_LOGLEVEL=warn NPM_CONFIG_PROGRESS=false
RUN yarn --pure-lockfile
COPY webpack-common.js webpack.local.config.js webpack.production.config.js server.js .babelrc .eslintrc .eslintignore .sass-lint.yml /home/weave/
