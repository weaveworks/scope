# NB only to be sourced

set -e

# these ought to match what is in Vagrantfile
# exported to override weave config.sh
export SSH_DIR="$PWD"
export HOSTS

: ${WEAVE_REPO:=github.com/weaveworks/weave}
: ${WEAVE_ROOT:="$(go list -e -f {{.Dir}} $WEAVE_REPO)"}

RUNNER="$WEAVE_ROOT/testing/runner/runner"
[ -x "$RUNNER" ] || (echo "Could not find weave test runner at $RUNNER." >&2 ; exit 1)

. "$WEAVE_ROOT/test/config.sh"

scope_on() {
  host=$1
  shift 1
  [ -z "$DEBUG" ] || greyly echo "Scope on $host: $@" >&2
  run_on $host DOCKER_HOST=tcp://$host:$DOCKER_PORT scope "$@"
}

weave_on() {
  host=$1
  shift 1
  [ -z "$DEBUG" ] || greyly echo "Weave on $host: $@" >&2
  run_on $host DOCKER_HOST=tcp://$host:$DOCKER_PORT weave "$@"
}
