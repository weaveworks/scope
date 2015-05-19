FROM mhart/alpine-node

WORKDIR /home/weave

# build tool
RUN npm install -g gulp

# install app and build dependencies
ADD package.json /home/weave/
RUN npm install

ADD gulpfile.js /home/weave/

# run container via
#
# `docker run -v $GOPATH/src/github.com/weaveworks/scope/client:/app weaveworks/scope-build`
#
# after the container is run, bundled app should be in ./dist/ dir
CMD gulp build
