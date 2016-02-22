#!/bin/bash

set -e

. ./config.sh

export PROJECT=scope-integration-tests
export TEMPLATE_NAME="test-template-5"
export NUM_HOSTS=3
. "../tools/integration/gce.sh" "$@"
