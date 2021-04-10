#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test host controls"

weave_on "$HOST1" launch
scope_on "$HOST1" launch

sleep 10

PROBEID=$(docker_on "$HOST1" logs weavescope 2>&1 | grep "probe starting" | sed -n 's/^.*ID \([0-9a-f]*\)$/\1/p')
HOSTID=$($SSH "$HOST1" hostname)

# Execute 'echo foo' in the host tty and check its output
PIPEID=$(curl -s -f -X POST "http://$HOST1:4040/api/control/$PROBEID/$HOSTID;<host>/host_exec" | jq -r '.pipe')
assert "(sleep 1 && echo \"PS1=''; echo foo\" && sleep 1) | wscat -b 'ws://$HOST1:4040/api/pipe/$PIPEID' | col -pb | tail -n 1" "foo\\n"

scope_end_suite
