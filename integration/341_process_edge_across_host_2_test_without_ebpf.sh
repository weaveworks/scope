#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test long connections (procspy) between processes on different hosts"

weave_on "$HOST1" launch "$HOST1" "$HOST2"
weave_on "$HOST2" launch "$HOST1" "$HOST2"

scope_on "$HOST1" launch --probe.ebpf.connections=false --probe.conntrack=false
scope_on "$HOST2" launch --probe.ebpf.connections=false --probe.conntrack=false

server_on "$HOST1"
weave_proxy_on "$HOST2" run -dti --name client alpine /bin/sh -c "while true; do \
	nc nginx.weave.local 80 || true; \
	sleep 1; \
done"

check() {
    wait_for processes "$1" 60 "nginx: worker process" nc
    has_connection processes "$1" nc "nginx: worker process"
    # Print connections and report in case of test failure
    list_connections "$1" processes
    echo "Report: $(curl -s "http://${1}:4040/api/report")"
}

check "$HOST1"
check "$HOST2"

scope_end_suite
