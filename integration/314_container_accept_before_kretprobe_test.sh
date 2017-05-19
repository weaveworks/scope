#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test accept before kretprobe, see https://github.com/weaveworks/tcptracer-bpf/issues/10"

weave_on "$HOST1" launch

# Launch the server before Scope to make sure it calls accept() before Scope's
# kretprobe on the accept function is installed. We use busybox' nc instead of
# Alpine's nc so that it blocks on the accept() syscall.
weave_on "$HOST1" run -d --name server busybox /bin/sh -c "while true; do \
		date ;
		sleep 1 ;
	done | nc -l -p 8080"

scope_on "$HOST1" launch --probe.ebpf.connections=true
wait_for_containers "$HOST1" 60 server
has_container "$HOST1" server

weave_on "$HOST1" run -d --name client busybox /bin/sh -c "ping -c 5 server.weave.local; \
	while true; do \
		date ;
		sleep 1 ;
	done | nc server.weave.local 8080"

wait_for_containers "$HOST1" 60 server client

has_container "$HOST1" client

list_containers "$HOST1"
list_connections "$HOST1"

has_connection containers "$HOST1" client server

endpoints_have_ebpf "$HOST1"

scope_end_suite
