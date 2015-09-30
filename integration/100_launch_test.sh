#! /bin/bash

. ./config.sh

start_suite "Launch scope and check it boots"

scope_on $HOST1 launch
assert_raises "curl $HOST1:4040"

end_suite
