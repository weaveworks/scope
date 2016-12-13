FROM node:6.9.0
WORKDIR /home/weave
COPY package.json /home/weave/
ENV NPM_CONFIG_LOGLEVEL=warn NPM_CONFIG_PROGRESS=false
RUN npm install --global yarn
RUN yarn
COPY webpack.local.config.js webpack.production.config.js server.js .babelrc .eslintrc .eslintignore /home/weave/
