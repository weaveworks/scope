#!/bin/bash

set -e

. ./config.sh

echo Copying scope images and scripts to hosts
for HOST in $HOSTS; do
	SIZE=$(stat --printf="%s" ../scope.tar)
	cat ../scope.tar | pv -N "scope.tar" -s $SIZE | $SSH -C $HOST sudo docker load
	upload_executable $HOST ../scope
	upload_executable $HOST ../scope /usr/local/scope/bin/scope
done

echo Installing weave
# Download the latest released weave script locally,
# for use by weave_on
curl -sL git.io/weave -o ./weave
chmod a+x ./weave
for HOST in $HOSTS; do
	run_on $HOST "sudo curl -sL git.io/weave -o /usr/local/bin/weave"
	run_on $HOST "sudo chmod a+x /usr/local/bin/weave"
	weave_on $HOST setup
done

echo Prefetching Images
for HOST in $HOSTS; do
	docker_on $HOST pull peterbourgon/tns-db
	docker_on $HOST pull alpine
	docker_on $HOST pull nginx
done

