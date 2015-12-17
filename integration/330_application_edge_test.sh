#! /bin/bash

. ./config.sh

start_suite "Test short connections between processes"

WEAVE_DOCKER_ARGS=$ADD_HOST_ARGS weave_on $HOST1 launch
scope_on $HOST1 launch
weave_on $HOST1 run -d --name nginx nginx
weave_on $HOST1 run -dti --name client alpine /bin/sh -c "while true; do \
	nc nginx.weave.local 80 || true; \
	sleep 1; \
done"

wait_for application $HOST1 60 nginx client

has applications $HOST1 nginx
has applications $HOST1 nc
has_connection applications $HOST1 nc nginx

scope_end_suite
