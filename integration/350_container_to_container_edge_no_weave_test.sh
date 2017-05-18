#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections between containers without Weave (no NAT)"

scope_on "$HOST1" launch
docker_on "$HOST1" run -d --name nginx nginx
wait_for_containers "$HOST1" 60 nginx
nginx_ip="$(docker_on "$HOST1" inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' nginx)"
docker_on "$HOST1" run -d --name client alpine /bin/sh -c "while true; do \
	wget $nginx_ip:80/ -O - >/dev/null || true; \
	sleep 1; \
done"
wait_for_containers "$HOST1" 60 client

has_container "$HOST1" nginx
has_container "$HOST1" client
has_connection containers "$HOST1" client nginx

scope_end_suite
