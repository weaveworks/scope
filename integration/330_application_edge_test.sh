#! /bin/bash

. ./config.sh

start_suite "Test long connections (procspy) between processes"

weave_on $HOST1 launch
scope_on $HOST1 launch --probe.conntrack=false
weave_on $HOST1 run -d --name nginx nginx
weave_on $HOST1 run -dti --name client alpine /bin/sh -c "while true; do \
	nc nginx.weave.local 80 || true; \
	sleep 1; \
done"

wait_for applications $HOST1 60 nginx client

has applications $HOST1 "nginx: worker process"
has applications $HOST1 nc
has_connection applications $HOST1 nc "nginx: worker process"

scope_end_suite
