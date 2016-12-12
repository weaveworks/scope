#! /bin/bash

. ./config.sh

start_suite "Launch 2 scopes and check they cluster (without weave)"

scope_on "$HOST1" launch --no-app "$HOST2"
scope_on "$HOST2" launch --no-probe

docker_on "$HOST1" run -dit --name db1 peterbourgon/tns-db

sleep 30 # need to allow the scopes to poll dns, resolve the other app ids, and send them reports.

has_container "$HOST2" weavescope
has_container "$HOST2" db1

scope_end_suite
