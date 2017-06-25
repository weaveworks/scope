#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections between containers"

weave_on "$HOST1" launch
scope_on "$HOST1" launch

server_on "$HOST1"
client_on "$HOST1"

wait_for_containers "$HOST1" 60 nginx client

has_container "$HOST1" nginx
has_container "$HOST1" client

list_containers "$HOST1"
list_connections "$HOST1"

has_connection containers "$HOST1" client nginx
endpoints_have_ebpf "$HOST1"

scope_end_suite
