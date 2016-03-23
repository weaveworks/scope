FROM node:4.2.2
WORKDIR /home/weave
COPY package.json /home/weave/
ENV NPM_CONFIG_LOGLEVEL=warn NPM_CONFIG_PROGRESS=false
# Don't install optional developer tools
RUN npm install --no-optional
COPY webpack.local.config.js webpack.production.config.js server.js .babelrc .eslintrc .eslintignore /home/weave/
