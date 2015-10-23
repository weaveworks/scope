#!/bin/bash

set -eux

# Mount the scope repo:
#  -v $(pwd):/go/src/github.com/weaveworks/scope

cd $GOPATH/src/github.com/weaveworks/scope
make app/scope-app
make probe/scope-probe

