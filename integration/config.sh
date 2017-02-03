#!/bin/bash
# NB only to be sourced

set -e

# these ought to match what is in Vagrantfile
# exported to override weave config.sh
export SSH_DIR="$PWD"
export HOSTS

# shellcheck disable=SC1091
. "../tools/integration/config.sh"

WEAVE="./weave"
SCOPE="../scope"

scope_on() {
    local host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Scope on $host: $*" >&2
    DOCKER_HOST=tcp://$host:$DOCKER_PORT CHECKPOINT_DISABLE=true "$SCOPE" "$@"
}

weave_on() {
    local host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Weave on $host: $*" >&2
    DOCKER_HOST=tcp://$host:$DOCKER_PORT CHECKPOINT_DISABLE=true "$WEAVE" "$@"
}

scope_end_suite() {
    end_suite
    for host in $HOSTS; do
        docker_on "$host" rm -f "$(docker_on "$host" ps -a -q)" 2>/dev/null 1>&2 || true
    done
}

# this checks we have a named node in the given view
has() {
    local view=$1
    local host=$2
    local name=$3
    local count=${4:-1}
    DEBUG=$(curl -s "http://${host}:4040/api/topology/${view}?system=all")
    assert "curl -s http://${host}:4040/api/topology/${view}?system=all | jq -r '[.nodes[] | select(.label == \"${name}\")] | length'" "$count"
}

# this checks we have a named container
has_container() {
    has containers "$@"
}

node_id() {
    local view="$1"
    local host="$2"
    local name="$3"
    curl -s "http://${host}:4040/api/topology/${view}?system=all" | jq -r ".nodes[] | select(.label == \"${name}\") | .id"
}

container_id() {
    node_id containers "$@"
}

# this checks we have an edge from container 1 to container 2
has_connection_by_id() {
    local view="$1"
    local host="$2"
    local from_id="$3"
    local to_id="$4"
    local timeout="${5:-60}"

    for i in $(seq "$timeout"); do
        local nodes
        local edge
        edge=$(echo "$nodes" | (jq -r ".nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])" || true) 2>/dev/null)
        nodes=$(curl -s "http://$host:4040/api/topology/${view}?system=all" || true)
        if [ "$edge" = "true" ]; then
            echo "Found edge $from -> $to after $i secs"
            assert "curl -s http://$host:4040/api/topology/${view}?system=all |  jq -r '.nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])'" true
            return
        fi
        sleep 1
    done

    echo "Failed to find edge $from -> $to after $timeout secs"
    assert "curl -s http://$host:4040/api/topology/${view}?system=all |  jq -r '.nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])'" true
}

has_connection() {
    local view="$1"
    local host="$2"
    local from="$3"
    local to="$4"
    local timeout="${5:-60}"
    local from_id
    local to_id

    from_id="$(node_id "${view}" "${host}" "${from}")"
    to_id="$(node_id "${view}" "${host}" "${to}")"

    has_connection_by_id "${view}" "${host}" "${from_id}" "${to_id}" "${timeout}"
}

wait_for() {
    local view="$1"
    local host="$2"
    local timeout="$3"
    shift 3

    for i in $(seq "${timeout}"); do
        local nodes
        local found=0
        nodes=$(curl -s "http://$host:4040/api/topology/${view}?system=all" || true)
        for name in "$@"; do
            local count
            count=$(echo "${nodes}" | jq -r "[.nodes[] | select(.label == \"${name}\")] | length")
            if [ -n "${count}" ] && [ "${count}" -ge 1 ]; then
                found=$((found + 1))
            fi
        done

        if [ "${found}" -eq $# ]; then
            echo "Found ${found} nodes after $i secs"
            return
        fi

        sleep 1
    done

    echo "Failed to find nodes $* after $i secs"
}

wait_for_containers() {
    wait_for containers "$@"
}
