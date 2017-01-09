#!/bin/bash

set -ex

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck disable=SC1090
. "$DIR/config.sh"

whitely echo Sanity checks
if ! bash "$DIR/sanity_check.sh"; then
    whitely echo ...failed
    exit 1
fi
whitely echo ...ok

# shellcheck disable=SC2068
TESTS=(${@:-$(find . -name '*_test.sh')})
RUNNER_ARGS=()

# If running on circle, use the scheduler to work out what tests to run
if [ -n "$CIRCLECI" ] && [ -z "$NO_SCHEDULER" ]; then
    RUNNER_ARGS=("${RUNNER_ARGS[@]}" -scheduler)
fi

# If running on circle or PARALLEL is not empty, run tests in parallel
if [ -n "$CIRCLECI" ] || [ -n "$PARALLEL" ]; then
    RUNNER_ARGS=("${RUNNER_ARGS[@]}" -parallel)
fi

make -C "${DIR}/../runner"
HOSTS="$HOSTS" "${DIR}/../runner/runner" "${RUNNER_ARGS[@]}" "${TESTS[@]}"
