# NB only to be sourced

set -e

# these ought to match what is in Vagrantfile
# exported to override weave config.sh
export SSH_DIR="$PWD"
export HOSTS

: ${WEAVE_REPO:=github.com/weaveworks/weave}
: ${WEAVE_ROOT:="$(go list -e -f {{.Dir}} $WEAVE_REPO)"}

WEAVE="./weave"
SCOPE="../scope"
RUNNER="$WEAVE_ROOT/testing/runner/runner"
[ -x "$RUNNER" ] || (echo "Could not find weave test runner at $RUNNER." >&2 ; exit 1)

. "$WEAVE_ROOT/test/config.sh"

scope_on() {
	local host=$1
	shift 1
	[ -z "$DEBUG" ] || greyly echo "Scope on $host: $@" >&2
	DOCKER_HOST=tcp://$host:$DOCKER_PORT $SCOPE "$@"
}

weave_on() {
	local host=$1
	shift 1
	[ -z "$DEBUG" ] || greyly echo "Weave on $host: $@" >&2
	DOCKER_HOST=tcp://$host:$DOCKER_PORT $WEAVE "$@"
}

# this checks we have a weavescope container
has_container() {
	local host=$1
	local name=$2
	local count=$3
	assert "curl -s http://$host:4040/api/topology/containers?system=show | jq -r '[.nodes | .[] | select(.label_major == \"$name\")] | length'" $count
}

scope_end_suite() {
	end_suite
	for host in $HOSTS; do
		docker_on $host rm -f $(docker_on $host ps -a -q) 2>/dev/null 1>&2 || true
	done
}
