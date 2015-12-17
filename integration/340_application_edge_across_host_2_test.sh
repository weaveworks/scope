#! /bin/bash

. ./config.sh

start_suite "Test connections between processes on different hosts"

WEAVE_DOCKER_ARGS=$ADD_HOST_ARGS weave_on $HOST1 launch $HOST1 $HOST2
WEAVE_DOCKER_ARGS=$ADD_HOST_ARGS weave_on $HOST2 launch $HOST1 $HOST2

scope_on $HOST1 launch
scope_on $HOST2 launch

weave_on $HOST1 run -d --name nginx nginx
weave_on $HOST2 run -dti --name client alpine /bin/sh -c "while true; do \
	nc nginx.weave.local 80 || true; \
	sleep 1; \
done"

sleep 30 # need to allow the scopes to poll dns, resolve the other app ids, and send them reports

check() {
    has applications $1 nginx
    has applications $1 nc
    has_connection applications $1 nc nginx
}

check $HOST1
check $HOST2

scope_end_suite
