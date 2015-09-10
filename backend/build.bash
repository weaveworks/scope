#!/bin/bash

mkdir -p /go/src/github.com/weaveworks
cd /go/src/github.com/weaveworks
git clone https://github.com/weaveworks/scope
cd scope
make deps
SUDO= make app/scope-app
mv app/scope-app $GOPATH/bin
SUDO= make probe/scope-probe
mv probe/scope-probe $GOPATH/bin

