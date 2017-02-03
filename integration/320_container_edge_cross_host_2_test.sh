#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test short lived connections between containers on different hosts"

weave_on "$HOST1" launch "$HOST1" "$HOST2"
weave_on "$HOST2" launch "$HOST1" "$HOST2"

scope_on "$HOST1" launch
scope_on "$HOST2" launch

weave_on "$HOST1" run -d --name nginx nginx
weave_on "$HOST2" run -d --name client alpine /bin/sh -c "while true; do \
	wget http://nginx.weave.local:80/ -O - >/dev/null || true; \
	sleep 1; \
done"

sleep 60 # need to allow the scopes to poll dns, resolve the other app ids, and send them reports

check() {
    has_container "$1" nginx
    has_container "$1" client
    has_connection containers "$1" client nginx
}

check "$HOST1"
check "$HOST2"

scope_end_suite
