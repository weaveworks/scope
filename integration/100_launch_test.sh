#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Launch scope and check it boots"

weave_on "$HOST1" launch
scope_on "$HOST1" launch

wait_for_containers "$HOST1" 60 weave weavescope

has_container "$HOST1" weave
has_container "$HOST1" weavescope

# Fail if the top-level UI is suspiciously small
ui_len="$(curl -s "http://$HOST1:4040/" | wc -c)"
assert_raises "(( $ui_len > 500 ))" 0

scope_end_suite
