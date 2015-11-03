#!/bin/bash

set -e

. ./config.sh

SCHEDULER_PREFIX=scope-integration
. $WEAVE_ROOT/test/run_all.sh "$@"
