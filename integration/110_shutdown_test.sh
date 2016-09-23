#! /bin/bash

. ./config.sh

start_suite "Check scope exits cleanly within 10 seconds"

scope_on $HOST1 launch
scope_on $HOST1 stop

sleep 10

# Save stdout for debugging output
exec 3>&1 
assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'app exiting' || (docker_on $HOST1 logs weavescope 2>&3 ; false)"
assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'probe exiting' || (docker_on $HOST1 logs weavescope 2>&3 ; false)"
assert_raises "docker_on $HOST1 inspect --format='{{.State.Running}}' weavescope" "false"

scope_end_suite
