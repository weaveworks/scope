#! /bin/bash

. ./config.sh

start_suite "Test short lived connections between containers"

WEAVE_NO_FASTDP=true weave_on $HOST1 launch
scope_on $HOST1 launch
weave_on $HOST1 run -d --name nginx nginx
weave_on $HOST1 run -d --name client alpine /bin/sh -c "while true; do \
	wget http://nginx.weave.local:80/ -O - >/dev/null || true; \
	sleep 1; \
done"

wait_for_containers $HOST1 60 nginx client

has_container $HOST1 nginx
has_container $HOST1 client
has_connection $HOST1 client nginx

scope_end_suite
