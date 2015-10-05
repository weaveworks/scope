#! /bin/bash

. ./config.sh

start_suite "Test short lived connections from the Internet"

weave_on $HOST1 launch
scope_on $HOST1 launch
docker_on $HOST1 run -d -p 80:80 --name nginx nginx

do_connections() {
	while true; do
		curl -s http://$HOST1:80/ >/dev/null
		sleep 1
	done
}
do_connections&

sleep 5 # give the probe a few seconds to build a report and send it to the app

has_container $HOST1 nginx 1
has_connection $HOST1 "The Internet" nginx

kill %do_connections

scope_end_suite
