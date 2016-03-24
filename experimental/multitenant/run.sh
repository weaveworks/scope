#!/bin/bash

set -eux

eval $(weave env)
CHECKPOINT_DISABLE=true

start_container() {
    local replicas=$1
    local image=$2
    local basename=$3
    shift 3
    local hostname=${basename}.weave.local

    local docker_args=
    while [ "$#" -gt 0 ]; do
        case "$1" in
            --)
            shift
            break
            ;;
            *)
            docker_args="${docker_args} $1"
            shift
            ;;
        esac
    done
    local container_args="$@"

    for i in $(seq ${replicas}); do
        if docker inspect ${basename}${i} >/dev/null 2>&1; then
            docker rm -f ${basename}${i}
        fi
        docker run -d -e CHECKPOINT_DISABLE --name=${basename}${i} --hostname=${hostname} \
            ${docker_args} ${image} ${container_args}
    done
}

# These are the infrastructure bits - do not use these containers in production
start_container 1 deangiberson/aws-dynamodb-local dynamodb
start_container 1 pakohan/elasticmq sqs
start_container 1 progrium/consul consul -p 8400:8400 -p 8500:8500 -p 8600:53/udp -- -server -bootstrap -ui-dir /ui

# These are the micro services
common_args="--no-probe --app.weave.addr= --app.http.address=:80"
start_container 2 weaveworks/scope collection -- ${common_args} \
                                                 --app.collector=dynamodb://abc:123@dynamodb.weave.local:8000 \
                                                 --app.aws.create.tables=true
start_container 2 weaveworks/scope query      -- ${common_args} \
                                                 --app.collector=dynamodb://abc:123@dynamodb.weave.local:8000
start_container 2 weaveworks/scope controls   -- ${common_args} \
                                                 --app.control.router=sqs://abc:123@sqs.weave.local:9324
start_container 2 weaveworks/scope pipes      -- ${common_args} \
                                                 --app.pipe.router=consul://consul.weave.local:8500/pipes/ \
                                                 --app.consul.inf=ethwe

# And we bring it all together with a reverse proxy
start_container 1 weaveworks/scope-frontend frontend --add-host=dns.weave.local:$(weave docker-bridge-ip) --publish=4040:80
