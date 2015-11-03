#!/bin/sh

set -eux

# Mount the scope repo:
#  -v $(pwd):/go/src/github.com/weaveworks/scope

cd $GOPATH/src/github.com/weaveworks/scope
make BUILD_IN_CONTAINER=false $@

