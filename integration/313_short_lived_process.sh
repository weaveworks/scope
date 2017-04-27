#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections in short lived processes between containers, with ebpf connection tracking enabled, require proc connector"

weave_on "$HOST1" launch
scope_on "$HOST1" launch --probe.ebpf.connections=true
weave_on "$HOST1" run -d --name server alpine /bin/sh -c "sleep 2 ; nc -l -p 8080 ; sleep 2"
weave_on "$HOST1" run -d --name client alpine /bin/sh -c "sleep 5 ; echo Hello | nc server.weave.local 8080 ; sleep 2"

wait_for_containers "$HOST1" 60 server client

has_container "$HOST1" server
has_container "$HOST1" client
has_connection containers "$HOST1" client server

endpoints_have_ebpf "$HOST1"

scope_end_suite
