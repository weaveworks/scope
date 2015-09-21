#!/bin/bash

function zombies() {
	if [ -z "$CIRCLECI" ]
	then
		docker exec $C ps -o pid,stat | grep Z | wc -l
	else
		# https://circleci.com/docs/docker#docker-exec
		sudo lxc-attach -n "$(docker inspect --format '{{.Id}}' $C)" -- sh -c "ps -o pid,stat | grep Z | wc -l"
	fi
}

function stop_rm() {
	docker stop $1
	docker rm $1
}

SLEEP=1
RC=0

C=$(docker run -d zombietest /runsvinit -reap=false)
sleep $SLEEP
NOREAP=$(zombies)
echo -n without reaping, we have $NOREAP zombies...
if [ "$NOREAP" -le "0" ]
then
	echo " FAIL"
	RC=1
else
	echo " good"
fi
stop_rm $C

C=$(docker run -d zombietest /runsvinit)
sleep $SLEEP
YESREAP=$(zombies)
echo -n with reaping, we have $YESREAP zombies...
if [ "$YESREAP" -gt "0" ]
then
	echo " FAIL"
	RC=1
else
	echo " good"
fi
stop_rm $C

exit $RC
