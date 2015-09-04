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

start_container 1 elasticsearch elasticsearch
start_container 2 tomwilkie/searchapp searchapp
start_container 1 redis redis
start_container 1 tomwilkie/qotd qotd
start_container 2 tomwilkie/app app
start_container 2 tomwilkie/frontend frontend --add-host=dns.weave.local:$(weave docker-bridge-ip)
start_container 1 tomwilkie/client client
