FROM debian:latest
WORKDIR /home/weave

RUN apt-get update && apt-get install -y curl bzip2 libfreetype6 libfontconfig1

# Install nodejs
RUN curl -sL https://deb.nodesource.com/setup | bash -
RUN apt-get install -y nodejs

# build tool
RUN npm install -g gulp

# install app and build dependencies
ADD package.json /home/weave/
RUN npm install

ADD gulpfile.js /home/weave/

# For instructions on running this container, consult the toplevel Makefile
