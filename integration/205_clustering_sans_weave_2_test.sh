#! /bin/bash

. ./config.sh

start_suite "Launch 2 scopes and check they cluster (without weave)"

scope_on $HOST1 launch $HOST1 $HOST2
scope_on $HOST2 launch $HOST1 $HOST2

docker_on $HOST1 run -dit --name db1 peterbourgon/tns-db
docker_on $HOST2 run -dit --name db2 peterbourgon/tns-db

sleep 30

check() {
	has_container $1 weave 0
	has_container $1 weaveproxy 0
	has_container $1 weavescope 2
	has_container $1 db1 1
	has_container $1 db2 1
}

check $HOST1
check $HOST2

scope_end_suite
