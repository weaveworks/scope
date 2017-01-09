#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections from the Internet"

if ! echo "$HOST1" | grep "us-central1-a"; then
	echo "Skipping; test needs to be run against VMs on GCE."
	scope_end_suite
	exit
fi

weave_on "$HOST1" launch
scope_on "$HOST1" launch
docker_on "$HOST1" run -d -p 80:80 --name nginx nginx

do_connections() {
	while true; do
		curl -s "http://$HOST1:80/" >/dev/null || true
		sleep 1
	done
}
do_connections&

wait_for_containers "$HOST1" 60 nginx "The Internet"

has_connection_by_id containers "$HOST1" "in-theinternet" "$(node_id containers "$HOST1" nginx)"

kill %do_connections

scope_end_suite
