#!/bin/bash

set -e # NB don't set -u, as weave's config.sh doesn't like that.

# shellcheck disable=SC1091
. ./config.sh

echo Copying scope images and scripts to hosts
# shellcheck disable=SC2153
for HOST in $HOSTS; do
    SIZE=$(stat --printf="%s" ../scope.tar)
    pv -N "scope.tar" -s "$SIZE" ../scope.tar | $SSH -C "$HOST" sudo docker load
done

setup_host() {
    local HOST=$1
    echo Installing weave on "$HOST"
    # Download the latest released weave script locally,
    # for use by weave_on
    curl -sL git.io/weave -o ./weave
    chmod a+x ./weave
    run_on "$HOST" "sudo curl -sL git.io/weave -o /usr/local/bin/weave"
    run_on "$HOST" "sudo chmod a+x /usr/local/bin/weave"
    weave_on "$HOST" setup

    echo Prefetching Images on "$HOST"
    docker_on "$HOST" pull peterbourgon/tns-db
    docker_on "$HOST" pull alpine
    docker_on "$HOST" pull nginx
}

for HOST in $HOSTS; do
    setup_host "$HOST" &
done

wait
