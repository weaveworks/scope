#! /bin/bash

. ./config.sh

start_suite "Launch 2 scopes and check they cluster"

weave_on $HOST2 launch
container_id=$(start_container $HOST2)

scope_on $HOST1 launch
scope_on $HOST2 launch
assert_raises "curl $HOST1:4040/api/topology/containers | grep '$container_id'"

end_suite
