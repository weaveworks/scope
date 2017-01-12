#!/bin/bash

set -ex

readonly ARG="$1"

if ! weave status 1>/dev/null 2>&1; then
    WEAVE_NO_PLUGIN=y weave launch
fi

eval "$(weave env)"

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
    local container_args="$*"

    for i in $(seq "${replicas}"); do
        if docker inspect "${basename}""${i}" >/dev/null 2>&1; then
            docker rm -f "${basename}""${i}"
        fi
        docker run -d -e CHECKPOINT_DISABLE --name="${basename}""${i}" --hostname="${hostname}" \
            "${docker_args}" "${image}" "${container_args}"
    done
}

start_container 1 tomwilkie/qotd qotd
start_container 1 tomwilkie/echo echo
start_container 1 tomwilkie/trace_app app
start_container 1 tomwilkie/client client -- -target app.weave.local \
    -concurrency 1 -persist False
