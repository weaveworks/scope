#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections from the Internet without ebpf [DISABLED]"

weave_on "$HOST1" launch
scope_on "$HOST1" launch --probe.ebpf.connections=false

## Test disabled: it is currently flaky
## https://github.com/weaveworks/scope/issues/2308

# docker_on "$HOST1" run -d -p 80:80 --name nginx nginx
#
# do_connections() {
#     while true; do
#         curl -s "http://$HOST1:80/" >/dev/null || true
#         sleep 1
#     done
# }
# do_connections &
#
# wait_for_containers "$HOST1" 60 nginx "The Internet"
#
# has_connection_by_id containers "$HOST1" "in-theinternet" "$(node_id containers "$HOST1" nginx)"
#
# endpoints_have_ebpf "$HOST1"
#
# kill %do_connections

scope_end_suite
