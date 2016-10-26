#! /bin/bash

. ./config.sh

start_suite "Launch scope and check it boots, with a spurious host arg"

scope_on "$HOST1" launch noatrealhost.foo

wait_for_containers "$HOST1" 60 weavescope

has_container "$HOST1" weavescope

scope_end_suite
