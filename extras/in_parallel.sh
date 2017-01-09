#!/bin/sh

set -e

if [ $# -lt 2 ]; then
    echo "Usage: $0 <cmd> [args...]"
    echo "  Will run cmd arg1, cmd arg2 etc on different circle shared,"
    echo "  based on what the scheduler says."
    exit 2
fi

if [ -z "$CIRCLECI" ]; then
    echo "I'm afraid this only works when run on CircleCI"
    exit 1
fi

COMMAND=$1
shift 1

INPUTS="$*"
SCHED_NAME=parallel-$CIRCLE_PROJECT_USERNAME-$CIRCLE_PROJECT_REPONAME-$CIRCLE_BUILD_NUM
INPUTS=$(echo "$INPUTS" | "../tools/sched" sched "$SCHED_NAME" "$CIRCLE_NODE_TOTAL" "$CIRCLE_NODE_INDEX")

echo Doing "$INPUTS"

for INPUT in $INPUTS; do
    START=$(date +%s)
    "$COMMAND" "$INPUT"
    RUNTIME=$(( $(date +%s) - START ))

    "../tools/sched" time "$INPUT" "$RUNTIME"
done
