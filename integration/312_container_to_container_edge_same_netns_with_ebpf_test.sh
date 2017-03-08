#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connection between containers in same network namespace, with ebpf connection tracking enabled"

scope_on "$HOST1" launch --probe.ebpf.connections=true
docker_on "$HOST1" run -d --name nginx nginx
docker_on "$HOST1" run -d --net=container:nginx --name client albanc/dialer /go/bin/dialer connectshortlived localhost:80

wait_for_containers "$HOST1" 60 nginx client

has_container "$HOST1" nginx
has_container "$HOST1" client
has_connection containers "$HOST1" client nginx

endpoints_have_ebpf "$HOST1"

scope_end_suite
