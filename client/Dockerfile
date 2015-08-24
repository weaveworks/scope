FROM node:0.10
WORKDIR /home/weave
COPY package.json /home/weave/
RUN npm install
COPY webpack.local.config.js webpack.production.config.js server.js .eslintrc /home/weave/
