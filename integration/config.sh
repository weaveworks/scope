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
        docker_on "$host" rm -f "$(docker_on "$host" ps -a -q)" || true
        # Unfortunately, "docker rm" might not work: the CircleCI's Docker
        # client is unable to delete containers on GCE's Docker server. As a
        # workaround, restart the Docker daemon: at least the containers from
        # previous tests will not be running.
        run_on "$host" "sudo service docker restart"
    done
}

list_containers() {
    local host=$1
    echo "Listing containers on ${host}:"
    curl -s "http://${host}:4040/api/topology/containers?system=show" | jq -r '.nodes[] | select(has("metadata")) | { "image": .metadata[] | select(.id == "docker_image_name") | .value, "label": .label, "id": .id} | .id + " (" + .image + ", " + .label + ")"'
    echo
}

list_connections() {
    local host=$1
    echo "Listing connections on ${host}:"
    curl -s "http://${host}:4040/api/topology/containers?system=show" | jq -r '.nodes[] | select(has("adjacency")) | { "from_name": .label, "from_id": .id, "to": .adjacency[]} | .from_id + " (" + .from_name+ ") -> " + .to'
    echo
}

# this checks we have a named node in the given view
has() {
    local view=$1
    local host=$2
    local name=$3
    local count=${4:-1}
    assert "curl -s http://${host}:4040/api/topology/${view}?system=show | jq -r '[.nodes[] | select(.label == \"${name}\")] | length'" "$count"
}

# this checks we have a named container
has_container() {
    has containers "$@"
}

node_id() {
    local view="$1"
    local host="$2"
    local name="$3"
    curl -s "http://${host}:4040/api/topology/${view}?system=show" | jq -r ".nodes[] | select(.label == \"${name}\") | .id"
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
        nodes=$(curl -s "http://$host:4040/api/topology/${view}?system=show" || true)
        if [ "$edge" = "true" ]; then
            echo "Found edge $from -> $to after $i secs"
            assert "curl -s http://$host:4040/api/topology/${view}?system=show |  jq -r '.nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])'" true
            return
        fi
        sleep 1
    done

    echo "Failed to find edge $from -> $to after $timeout secs"
    assert "curl -s http://$host:4040/api/topology/${view}?system=show |  jq -r '.nodes[\"$from_id\"].adjacency | contains([\"$to_id\"])'" true
}

# this checks if ebpf is true on all endpoints on a given host
endpoints_have_ebpf() {
    local host="$1"
    local timeout="${2:-60}"
    local number_of_endpoints=-1
    local have_ebpf=-1
    local report

    for i in $(seq "$timeout"); do
        report=$(curl -s "http://${host}:4040/api/report")
        number_of_endpoints=$(echo "${report}" | jq -r '.Endpoint.nodes | length')
        have_ebpf=$(echo "${report}" | jq -r '.Endpoint.nodes[].latest.eBPF | select(.value != null) | contains({"value": "true"})' | wc -l)
        if [[ "$number_of_endpoints" -gt 0 && "$have_ebpf" -gt 0 && "$number_of_endpoints" -eq "$have_ebpf" ]]; then
            echo "Found ${number_of_endpoints} endpoints with ebpf enabled"
            assert "echo '$have_ebpf'" "$number_of_endpoints"
            return
        fi
        sleep 1
    done

    echo "Only ${have_ebpf} endpoints of ${number_of_endpoints} have ebpf enabled, should be equal"
    echo "Example of one endpoint:"
    echo "${report}" | jq -r '[.Endpoint.nodes[]][0]'
    assert "echo '$have_ebpf" "$number_of_endpoints"
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
        nodes=$(curl -s "http://$host:4040/api/topology/${view}?system=show" || true)
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
