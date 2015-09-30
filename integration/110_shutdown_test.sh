#! /bin/bash

. ./config.sh

start_suite "Check scope exits cleanly within 10 seconds"

scope_on $HOST1 launch
scope_on $HOST1 stop

sleep 10

assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'app exiting'"
assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'probe exiting'"
assert_raises "docker_on $HOST1 inspect --format='{{.State.Running}}' weavescope" "false"

scope_end_suite
