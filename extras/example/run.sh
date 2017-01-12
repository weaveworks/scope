#!/bin/bash

set -ex

readonly ARG="$1"

if ! weave status 1>/dev/null 2>&1; then
    WEAVE_NO_PLUGIN=y weave launch
fi

eval "$(weave env)"

start_container() {
    local IMAGE=$2
    local BASENAME=$3
    local REPLICAS=$1
    shift 3
    local HOSTNAME=$BASENAME.weave.local

    for i in $(seq "$REPLICAS"); do
        if docker inspect "$BASENAME""$i" >/dev/null 2>&1; then
            docker rm -f "$BASENAME""$i"
        fi
        if [ "$ARG" != "-rm" ]; then
            docker run -d --name="$BASENAME""$i" --hostname="$HOSTNAME" "$@" "$IMAGE"
        fi
    done
}

start_container 1 elasticsearch elasticsearch
start_container 2 tomwilkie/searchapp search
start_container 1 redis redis
start_container 1 tomwilkie/qotd qotd
start_container 1 tomwilkie/echo echo
start_container 2 tomwilkie/app app
start_container 2 tomwilkie/frontend frontend --add-host=dns.weave.local:"$(weave docker-bridge-ip)"
start_container 1 tomwilkie/client client
