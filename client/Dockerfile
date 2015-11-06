FROM node:0.10
WORKDIR /home/weave
COPY package.json /home/weave/
# Dont install optional developer tools
RUN npm install --no-optional
COPY webpack.local.config.js webpack.production.config.js server.js .eslintrc .eslintignore /home/weave/
