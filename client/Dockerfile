FROM node:6.9.0
WORKDIR /home/weave
COPY package.json yarn.lock /home/weave/
ENV NPM_CONFIG_LOGLEVEL=warn NPM_CONFIG_PROGRESS=false
RUN curl -o- -L https://yarnpkg.com/install.sh | bash
RUN yarn
COPY webpack.local.config.js webpack.production.config.js server.js .babelrc .eslintrc .eslintignore /home/weave/
