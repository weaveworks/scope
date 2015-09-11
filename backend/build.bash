#!/bin/bash

set -x

# Mount the scope repo:
#  -v $(pwd):/go/src/github.com/weaveworks/scope

cd /go/src/github.com/weaveworks/scope
make deps
make app/scope-app
make probe/scope-probe

