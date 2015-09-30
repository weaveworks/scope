#! /bin/bash

. ./config.sh

start_suite "Launch scope and check it boots"

weave_on $HOST1 launch
scope_on $HOST1 launch

sleep 5 # give the probe a few seconds to build a report and send it to the app

has_container $HOST1 weave 1
has_container $HOST1 weaveproxy 1
has_container $HOST1 weavescope 1

scope_end_suite
