#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test container controls"

weave_on "$HOST1" launch
scope_on "$HOST1" launch

CID=$(weave_proxy_on "$HOST1" run -dti --name alpine -e PATH=/home:/usr/bin alpine /bin/sh)

wait_for_containers "$HOST1" 60 alpine

assert "docker_on $HOST1 inspect --format='{{.State.Running}}' alpine" "true"
PROBEID=$(docker_on "$HOST1" logs weavescope 2>&1 | grep "probe starting" | sed -n 's/^.*ID \([0-9a-f]*\)$/\1/p')

# Execute 'echo $PATH' in a container tty and check its output - as
# well as checking basic operation, this also checks that the
# container's PATH settings are respected, which isn't the case for
# login shells.
PIPEID=$(curl -s -f -X POST "http://$HOST1:4040/api/control/$PROBEID/$CID;<container>/docker_exec_container" | jq -r '.pipe')
assert "(sleep 1 && echo 'echo \$PATH' && sleep 1) | wscat -b 'ws://$HOST1:4040/api/pipe/$PIPEID' | col -pb" "/ # 6necho \$PATH\\n/home:/usr/bin\\n/ # 6n"

assert_raises "curl -f -X POST  'http://$HOST1:4040/api/control/$PROBEID/$CID;<container>/docker_stop_container'"

sleep 5
assert "docker_on $HOST1 inspect --format='{{.State.Running}}' alpine" "false"

scope_end_suite
