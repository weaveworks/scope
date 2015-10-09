#! /bin/bash

. ./config.sh

start_suite "Test short lived connections from the Internet"

weave_on $HOST1 launch
scope_on $HOST1 launch
docker_on $HOST1 run -d -p 80:80 --name nginx nginx

do_connections() {
	while true; do
		curl -s http://$HOST1:80/ >/dev/null || true
		sleep 1
	done
}
do_connections&

wait_for_containers $HOST1 60 nginx "The Internet"

has_container $HOST1 nginx
has_container $HOST1 "The Internet"
has_connection $HOST1 "The Internet" nginx 60

kill %do_connections

scope_end_suite
