#!/bin/bash

set -e

. ./config.sh

NO_SCHEDULER=1 ../tools/integration/run_all.sh "$@"
