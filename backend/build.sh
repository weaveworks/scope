#!/bin/sh

set -eu

SCOPE_SRC=$GOPATH/src/github.com/weaveworks/scope

# Mount the scope repo:
#  -v $(pwd):/go/src/github.com/weaveworks/scope

make -C $SCOPE_SRC BUILD_IN_CONTAINER=false $*
