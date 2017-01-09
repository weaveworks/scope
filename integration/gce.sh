#!/bin/bash

set -e

# shellcheck disable=SC1091
. ./config.sh

export PROJECT=scope-integration-tests
export TEMPLATE_NAME="test-template-5"
export NUM_HOSTS=5
# shellcheck disable=SC1091
. "../tools/integration/gce.sh" "$@"
