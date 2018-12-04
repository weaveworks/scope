#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test with ebpf restarts and proc fallback"

weave_on "$HOST1" launch
# Manually start scope in order to start EbpfTracker in debug mode
DOCKER_HOST=tcp://${HOST1}:${DOCKER_PORT} CHECKPOINT_DISABLE=true \
    WEAVESCOPE_DOCKER_ARGS="-e SCOPE_DEBUG_BPF=1" \
    "${SCOPE}" launch

server_on "$HOST1"
client_on "$HOST1"

wait_for_containers "$HOST1" 60 nginx client

has_container "$HOST1" nginx
has_container "$HOST1" client
has_connection containers "$HOST1" client nginx

docker_on "$HOST1" exec weavescope sh -c "echo stop > /var/run/scope/debug-bpf"
sleep 5

server_on "$HOST1" "nginx2"
client_on "$HOST1" "client2" "nginx2"

wait_for_containers "$HOST1" 60 nginx2 client2

has_container "$HOST1" nginx2
has_container "$HOST1" client2
has_connection containers "$HOST1" client2 nginx2

# Save stdout for debugging output
exec 3>&1
assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'ebpf tracker died, restarting it' || (docker_on $HOST1 logs weavescope 2>&3 ; false)"

docker_on "$HOST1" exec weavescope sh -c "echo stop > /var/run/scope/debug-bpf"
sleep 5

server_on "$HOST1" "nginx3"
client_on "$HOST1" "client3" "nginx3"

wait_for_containers "$HOST1" 60 nginx3 client3

has_container "$HOST1" nginx3
has_container "$HOST1" client3
has_connection containers "$HOST1" client3 nginx3

assert_raises "docker_on $HOST1 logs weavescope 2>&1 | grep 'ebpf tracker died again, gently falling back to proc scanning' || (docker_on $HOST1 logs weavescope 2>&3 ; false)"

scope_end_suite
