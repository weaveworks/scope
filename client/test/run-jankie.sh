#!/bin/bash
# run jankie on one commit

set -x

HOST=$1
COMMIT=$2
DATE=$3

git checkout $COMMIT
make SUDO= -C ../..
../../scope stop && ../../scope launch

sleep 5

COMMIT=$COMMIT DATE=$DATE HOST=$HOST DEBUG=scope* node ./perfjankie/main.js
