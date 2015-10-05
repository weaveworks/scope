#! /bin/bash

. ./config.sh

start_suite "Test short lived connections between containers"

weave_on $HOST1 launch
scope_on $HOST1 launch
weave_on $HOST1 run -d --name nginx nginx
weave_on $HOST1 run -d --name client alpine /bin/sh -c "while true; do wget http://nginx.weave.local:80/ >/dev/null; sleep 1; done"

sleep 20 # give the probe a few seconds to build a report and send it to the app

has_container $HOST1 nginx 1
has_container $HOST1 client 1
has_connection $HOST1 client nginx

scope_end_suite
