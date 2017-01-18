#!/bin/bash

set -e

# shellcheck disable=SC1091
. ./config.sh

../tools/integration/run_all.sh "$@"
