FROM node:0.10
WORKDIR /home/weave
RUN npm install -g gulp
COPY package.json /home/weave/
RUN npm install
COPY gulpfile.js webpack.config.js .eslintrc /home/weave/
