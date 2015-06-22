# NB only to be sourced

set -e

# these ought to match what is in Vagrantfile
# exported to override weave config.sh
export SSH_DIR="$PWD"
export HOSTS

WEAVE_REPO="github.com/weaveworks/weave"
WEAVE_ROOT="${GOPATH%%:*}/src/$WEAVE_REPO"
if [ ! -d "$WEAVE_ROOT" ] ; then
  mkdir -p "$(dirname "$WEAVE_ROOT")"
  git clone --depth 1 -b master https://$WEAVE_REPO.git "$WEAVE_ROOT"
  go get $WEAVE_REPO/...
fi

. "$WEAVE_ROOT/test/config.sh"

scope_on() {
  host=$1
  shift 1
  [ -z "$DEBUG" ] || greyly echo "Scope on $host: $@" >&2
  run_on $host "DOCKER_HOST=tcp://$host:$DOCKER_PORT" scope "$@"
}
