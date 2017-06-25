#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Launch scope (without weave installed) and check it boots"

scope_on "$HOST1" launch

wait_for_containers "$HOST1" 60 weavescope

has_container "$HOST1" weave 0
has_container "$HOST1" weavescope

scope_end_suite
