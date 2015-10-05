#!/bin/bash

set -e

. ./config.sh

echo Copying scope images and scripts to hosts
for HOST in $HOSTS; do
    docker_on $HOST load -i ../scope.tar
    upload_executable $HOST ../scope
    upload_executable $HOST ../scope /usr/local/scope/bin/scope
done

echo Installing weave
for HOST in $HOSTS; do
    run_on $HOST "sudo curl -sL git.io/weave -o /usr/local/bin/weave"
    run_on $HOST "sudo chmod a+x /usr/local/bin/weave"
done

echo Prefetching Images
for HOST in $HOSTS; do
    weave_on $HOST setup
    docker_on $HOST pull peterbourgon/tns-db
    docker_on $HOST pull alpine
done

curl -sL git.io/weave -o ./weave
