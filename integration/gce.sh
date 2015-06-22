#!/bin/bash

set -e

. ./config.sh

export PROJECT=scope-integration-tests
. "$WEAVE_ROOT/test/gce.sh" "$@"
