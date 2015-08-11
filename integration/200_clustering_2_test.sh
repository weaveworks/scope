#! /bin/bash

. ./config.sh

start_suite "Launch 2 scopes and check they cluster"

weave_on $HOST2 launch
weave_on $HOST2 launch-dns
docker_on $HOST2 run -dit --name db1 peterbourgon/tns-db
container_id=$(docker_on $HOST2 run -dit --name app1 --link db1:db1 peterbourgon/tns-app)

scope_on $HOST1 launch $HOST2
scope_on $HOST2 launch $HOST1

SUCCESS=
for i in {1..10}; do
  if (curl -s $HOST1:4040/api/topology/containers | grep "$container_id" >/dev/null); then
    SUCCESS=1
    break
  fi
  sleep 1
done
assert "echo $SUCCESS" "1"

end_suite
