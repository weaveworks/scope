#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test long connections (procspy) between processes"

weave_on "$HOST1" launch
scope_on "$HOST1" launch --probe.conntrack=false

server_on "$HOST1"
weave_proxy_on "$HOST1" run -dti --name client alpine /bin/sh -c "while true; do \
	nc nginx.weave.local 80 || true; \
	sleep 1; \
done"

wait_for processes "$HOST1" 60 "nginx: worker process" nc

has processes "$HOST1" "nginx: worker process"
has processes "$HOST1" nc
has_connection processes "$HOST1" nc "nginx: worker process"

scope_end_suite
