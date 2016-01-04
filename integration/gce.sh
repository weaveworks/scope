#!/bin/bash

set -e

. ./config.sh

export PROJECT=scope-integration-tests
export TEMPLATE_NAME="test-template-4"
export NUM_HOSTS=2
. "$WEAVE_ROOT/test/gce.sh" "$@"
