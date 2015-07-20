#! /bin/bash

. ./config.sh

start_suite "Launch scope (without weave installed) and check it boots"

assert_raises "run_on $HOST1 \
  PATH=/usr/local/scope/bin:/usr/sbin:/usr/bin:/sbin:/bin \
  DOCKER_HOST=tcp://$HOST1:$DOCKER_PORT \
  scope launch"

assert_raises "curl $HOST1:4040"

end_suite
