#! /bin/bash

. ./config.sh

start_suite "Test container controls"

weave_on $HOST1 launch
scope_on $HOST1 launch

CID=$(weave_on $HOST1 run -dti --name alpine alpine /bin/sh)

wait_for_containers $HOST1 60 alpine

assert "docker_on $HOST1 inspect --format='{{.State.Running}}' alpine" "true"
PROBEID=$(docker_on $HOST1 logs weavescope 2>&1 | grep "probe starting" | sed -n 's/^.*ID \([0-9a-f]*\)$/\1/p')

# Execute 'echo foo' in a container tty and check its output
PIPEID=$(curl -s -f -X POST  "http://$HOST1:4040/api/control/$PROBEID/$CID;<container>/docker_exec_container" | jq -r '.pipe' )
assert "(sleep 1 && echo 'echo foo' && sleep 1) | wscat -b 'ws://$HOST1:4040/api/pipe/$PIPEID' | col -pb" "alpine:/# 6necho foo\nfoo\nalpine:/# 6n"

assert_raises "curl -f -X POST  'http://$HOST1:4040/api/control/$PROBEID/$CID;<container>/docker_stop_container'"

sleep 5
assert "docker_on $HOST1 inspect --format='{{.State.Running}}' alpine" "false"

scope_end_suite
