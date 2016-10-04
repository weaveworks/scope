FROM node:4.2.2
WORKDIR /home/weave
COPY package.json /home/weave/
ENV NPM_CONFIG_LOGLEVEL=warn NPM_CONFIG_PROGRESS=false
RUN npm install
COPY webpack.local.config.js webpack.production.config.js webpack.external.config.js server.js .babelrc .eslintrc .eslintignore /home/weave/
