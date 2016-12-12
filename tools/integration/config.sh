#!/bin/bash
# NB only to be sourced

set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Protect against being sourced multiple times to prevent
# overwriting assert.sh global state
if ! [ -z "$SOURCED_CONFIG_SH" ]; then
    return
fi
SOURCED_CONFIG_SH=true

# these ought to match what is in Vagrantfile
N_MACHINES=${N_MACHINES:-3}
IP_PREFIX=${IP_PREFIX:-192.168.48}
IP_SUFFIX_BASE=${IP_SUFFIX_BASE:-10}

if [ -z "$HOSTS" ]; then
    for i in $(seq 1 "$N_MACHINES"); do
        IP="${IP_PREFIX}.$((IP_SUFFIX_BASE + i))"
        HOSTS="$HOSTS $IP"
    done
fi

# these are used by the tests
# shellcheck disable=SC2034
HOST1=$(echo "$HOSTS" | cut -f 1 -d ' ')
# shellcheck disable=SC2034
HOST2=$(echo "$HOSTS" | cut -f 2 -d ' ')
# shellcheck disable=SC2034
HOST3=$(echo "$HOSTS" | cut -f 3 -d ' ')

# shellcheck disable=SC1090
. "$DIR/assert.sh"

SSH_DIR=${SSH_DIR:-$DIR}
sshcmd() {
  # shellcheck disable=SC2029
  ssh -l vagrant \
    -i "$SSH_DIR/insecure_private_key" \
    -o "UserKnownHostsFile=$SSH_DIR/.ssh_known_hosts" \
    -o CheckHostIP=no \
    -o StrictHostKeyChecking=no \
    "$@"
}
SMALL_IMAGE="alpine"
# shellcheck disable=SC2034
TEST_IMAGES="$SMALL_IMAGE"

# shellcheck disable=SC2034
PING="ping -nq -W 1 -c 1"
DOCKER_PORT=2375

remote() {
    rem=$1
    shift 1
    "$@" > >(while read -r line; do echo -e $'\e[0;34m'"$rem>"$'\e[0m'" $line"; done)
}

colourise() {
    ([ -t 0 ] && echo -ne $'\e['"$1"'m') || true
    shift
    # It's important that we don't do this in a subshell, as some
    # commands we execute need to modify global state
    "$@"
    ([ -t 0 ] && echo -ne $'\e[0m') || true
}

whitely() {
    colourise '1;37' "$@"
}

greyly() {
    colourise '0;37' "$@"
}

redly() {
    colourise '1;31' "$@"
}

greenly() {
    colourise '1;32' "$@"
}

run_on() {
    host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Running on $host:" "$@" >&2
    remote "$host" "ssh" "$host" "$@"
}

docker_on() {
    host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Docker on $host:$DOCKER_PORT:" "$@" >&2
    docker -H "tcp://$host:$DOCKER_PORT" "$@"
}

weave_on() {
    host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Weave on $host:$DOCKER_PORT:" "$@" >&2
    DOCKER_HOST=tcp://$host:$DOCKER_PORT $WEAVE "$@"
}

exec_on() {
    host=$1
    container=$2
    shift 2
    docker -H "tcp://$host:$DOCKER_PORT" exec "$container" "$@"
}

rm_containers() {
    host=$1
    shift
    [ $# -eq 0 ] || docker_on "$host" rm -f "$@" >/dev/null
}

start_suite() {
    for host in $HOSTS; do
        [ -z "$DEBUG" ] || echo "Cleaning up on $host: removing all containers and resetting weave"
        PLUGIN_ID=$(docker_on "$host" ps -aq --filter=name=weaveplugin)
        PLUGIN_FILTER="cat"
        [ -n "$PLUGIN_ID" ] && PLUGIN_FILTER="grep -v $PLUGIN_ID"
        rm_containers "$host" "$(docker_on "$host" ps -aq 2>/dev/null | "$PLUGIN_FILTER")"
        run_on "$host" "docker network ls | grep -q ' weave ' && docker network rm weave" || true
        weave_on "$host" reset 2>/dev/null
    done
    whitely echo "$@"
}

end_suite() {
    whitely assert_end
}

WEAVE=$DIR/../weave
