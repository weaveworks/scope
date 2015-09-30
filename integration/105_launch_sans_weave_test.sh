#! /bin/bash

. ./config.sh

start_suite "Launch scope (without weave installed) and check it boots"

scope_on $HOST1 launch

sleep 5 # give the probe a few seconds to build a report and send it to the app

has_container $HOST1 weave 0
has_container $HOST1 weaveproxy 0
has_container $HOST1 weavescope 1

scope_end_suite
