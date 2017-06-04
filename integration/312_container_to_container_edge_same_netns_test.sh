#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connection between containers in same network namespace"

scope_on "$HOST1" launch
docker_on "$HOST1" run -d --name nginx nginx
docker_on "$HOST1" run -d --net=container:nginx --name client albanc/dialer /go/bin/dialer connectshortlived localhost:80

wait_for_containers "$HOST1" 60 nginx client

has_container "$HOST1" nginx
has_container "$HOST1" client

list_containers "$HOST1"
list_connections "$HOST1"

has_connection containers "$HOST1" client nginx

endpoints_have_ebpf "$HOST1"

scope_end_suite
