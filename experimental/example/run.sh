#!/bin/bash

set -eux

REPLICAS=${REPLICAS:-1}

start_container() {
    IMAGE=$1
    BASENAME=${2:-$1}
    HOSTNAME=$BASENAME.weave.local

    for i in $(seq $REPLICAS); do
        if docker inspect $BASENAME$i >/dev/null 2>&1; then
            docker rm -f $BASENAME$i
        fi
        weave run --with-dns --name=$BASENAME$i --hostname=$HOSTNAME $IMAGE
    done
}

start_container elasticsearch
start_container tomwilkie/searchapp searchapp
start_container redis
start_container tomwilkie/qotd qotd
start_container tomwilkie/app app
start_container tomwilkie/client client
