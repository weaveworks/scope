#!/bin/bash
# This scripts copies all the coverage reports from various circle shards,
# merges them and produces a complete report.

set -ex
DESTINATION=$1
FROMDIR=$2
mkdir -p "$DESTINATION"

if [ -n "$CIRCLECI" ]; then
    for i in $(seq 1 $((CIRCLE_NODE_TOTAL - 1))); do
        scp "node$i:$FROMDIR"/* "$DESTINATION" || true
    done
fi

go get github.com/weaveworks/build-tools/cover
cover "$DESTINATION"/* >profile.cov
go tool cover -html=profile.cov -o coverage.html
go tool cover -func=profile.cov -o coverage.txt
tar czf coverage.tar.gz "$DESTINATION"
