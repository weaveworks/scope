FROM node:0.10
WORKDIR /home/weave
COPY package.json gulpfile.js webpack.config.js .eslintrc /home/weave/
RUN npm install -g gulp
RUN npm install
