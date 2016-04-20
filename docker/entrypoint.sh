#!/bin/bash

usage() {
	echo "$0 --app.foo bar --probe.foo bar"
	exit 1
}

mkdir -p /var/run/weave

TOKEN_PROVIDED=false

if [ "$1" = "version" -o "$1" = "help" ]; then
    exec -a scope /home/weave/scope --mode $1
    exit 0
fi

for arg in $@; do
    case "$arg" in
        --no-app|--probe-only)
            touch /etc/service/app/down
            ;;
        --no-probe|--app-only)
            touch /etc/service/probe/down
            ;;
        --service-token*)
            TOKEN_PROVIDED=true
            touch /etc/service/app/down
            ;;
    esac
done

echo "$@" >/var/run/weave/scope-app.args
echo "$@" >/var/run/weave/scope-probe.args

# End of the command line can optionally be some
# addresses of apps to connect to, for people not
# using Weave DNS. We stick these in /var/run/weave/apps
# for the run-probe script to pick up.
MANUAL_APPS=""

# Implicitly target the Scope Service if a service token was provided with
# no explicit manual app.
if [ "$MANUAL_APPS" = "" -a "$TOKEN_PROVIDED" = "true" ]; then
    MANUAL_APPS="scope.weave.works:443"
fi

echo "$MANUAL_APPS" >/var/run/weave/apps

exec /home/weave/runsvinit
