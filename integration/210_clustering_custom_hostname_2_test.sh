#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Launch 2 scopes and check they cluster automatically, with custom weave domain"

weave_on "$HOST1" launch --dns-domain foo.local "$HOST1" "$HOST2"
weave_on "$HOST2" launch --dns-domain foo.local "$HOST1" "$HOST2"

scope_on "$HOST1" launch --weave.hostname=bar.foo.local
scope_on "$HOST2" launch --weave.hostname bar.foo.local

docker_on "$HOST1" run -dit --name db1 peterbourgon/tns-db
docker_on "$HOST2" run -dit --name db2 peterbourgon/tns-db

sleep 30 # need to allow the scopes to poll dns, resolve the other app ids, and send them reports

check() {
	has_container "$1" weave 2
	has_container "$1" weaveproxy 2
	has_container "$1" weavescope 2
	has_container "$1" db1
	has_container "$1" db2
}

check "$HOST1"
check "$HOST2"

scope_end_suite
