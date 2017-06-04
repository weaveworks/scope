#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections between containers, with ebpf proc fallback"

weave_on "$HOST1" launch
# Manually start scope in order to set
#    `WEAVESCOPE_DOCKER_ARGS="-v /tmp:/sys/kernel/debug/tracing:ro"`
# to make ebpf fail and test the proc fallback.
DOCKER_HOST=tcp://${HOST1}:${DOCKER_PORT} CHECKPOINT_DISABLE=true \
    WEAVESCOPE_DOCKER_ARGS="-v /tmp:/sys/kernel/debug/tracing:ro" \
    "${SCOPE}" launch
weave_on "$HOST1" run -d --name nginx nginx
weave_on "$HOST1" run -d --name client alpine /bin/sh -c "while true; do \
	wget http://nginx.weave.local:80/ -O - >/dev/null || true; \
	sleep 1; \
done"

wait_for_containers "$HOST1" 60 nginx client

has_container "$HOST1" nginx
has_container "$HOST1" client
has_connection containers "$HOST1" client nginx

# Save stdout for debugging output
exec 3>&1
assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'Error setting up the eBPF tracker, falling back to proc scanning' || (docker_on $HOST1 logs weavescope 2>&3 ; false)"

scope_end_suite
