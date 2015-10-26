#!/bin/sh

set -eux

# Mount the scope repo:
#  -v $(pwd):/go/src/github.com/weaveworks/scope

cd $GOPATH/src/github.com/weaveworks/scope
rm $1 2>/dev/null || true
make LOCAL=true $1

