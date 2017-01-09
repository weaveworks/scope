#!/bin/bash
# run jankie on one commit

# These need to be running:
#
# docker run --net=host -d -p 4444:4444 -v /dev/shm:/dev/shm selenium/standalone-chrome:2.52.0
# docker run -d -p 5984:5984 --name couchdb klaemo/couchdb
#
# Initialize the results DB
#
# perfjankie --only-update-site  --couch-server http://local.docker:5984 --couch-database performance
#
# Optional: run from localhost which can be a bit fast than rebuilding...
#  - ssh -R 0.0.0.0:4042:localhost:4042 docker@local.docker
#  - npm run build
#  - BACKEND_HOST=local.docker npm start
#  - ./run-jankie.sh 192.168.64.3:4042
#
# Usage:
#
# ./run-jankie.sh 192.168.64.3:4040
#
# View results: http://local.docker:5984/performance/_design/site/index.html
#

set -x
set -e

HOST="$1"
DATE=$(git log --format="%at" -1)
COMMIT=$(git log --format="%h" -1)

echo "Testing $COMMIT on $DATE"

# ../../scope stop
# make SUDO= -C ../..
# ../../scope launch
# sleep 5

COMMIT="$COMMIT" DATE=$DATE HOST=$HOST DEBUG="scope*" node ./perfjankie/main.js
