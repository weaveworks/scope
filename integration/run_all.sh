#!/bin/bash

set -e

. ./config.sh

NO_SCHEDULER=1 $WEAVE_ROOT/test/run_all.sh "$@"
