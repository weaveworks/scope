#!/bin/bash

set -eux

eval $(weave env)

start_container() {
    IMAGE=$2
    BASENAME=$3
    REPLICAS=$1
    shift 3
    HOSTNAME=$BASENAME.weave.local

    for i in $(seq $REPLICAS); do
        if docker inspect $BASENAME$i >/dev/null 2>&1; then
            docker rm -f $BASENAME$i
        fi
        docker run -d --name=$BASENAME$i --hostname=$HOSTNAME $@ $IMAGE
    done
}

start_container 1 tomwilkie/qotd qotd
start_container 1 tomwilkie/app app

