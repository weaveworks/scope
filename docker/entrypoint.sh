#!/bin/sh

usage() {
	echo "$0 --app.foo bar --probe.foo bar"
	exit 1
}

mkdir -p /etc/weave
APP_ARGS=""
PROBE_ARGS=""
TOKEN_PROVIDED=false

if [ "$1" = version ]; then
    /home/weave/scope version
    exit 0
fi

while true; do
    case "$1" in
        --app.*)
            if echo "$1" | grep "=" 1>/dev/null; then
                ARG_NAME=$(echo "$1" | sed 's/\-\-app\.\([^=]*\)=\(.*\)/\1/')
                ARG_VALUE=$(echo "$1" | sed 's/\-\-app\.\([^=]*\)=\(.*\)/\2/')
            else
                [ $# -gt 1 ] || usage
                ARG_NAME=$(echo "$1" | sed 's/\-\-app\.//')
                ARG_VALUE="$2"
                shift
            fi
            APP_ARGS="$APP_ARGS -$ARG_NAME=$ARG_VALUE"
            ;;
        --debug)
            APP_ARGS="$APP_ARGS -log.level=debug"
            PROBE_ARGS="$PROBE_ARGS -log.level=debug"
            ;;
        --no-app|--app-only)
            touch /etc/service/app/down
            ;;
        --no-probe|--probe-only)
            touch /etc/service/probe/down
            ;;
        --probe.*)
            if echo "$1" | grep "=" 1>/dev/null; then
                ARG_NAME=$(echo "$1" | sed 's/\-\-probe\.\([^=]*\)=\(.*\)/\1/')
                ARG_VALUE=$(echo "$1" | sed 's/\-\-probe\.\([^=]*\)=\(.*\)/\2/')
            else
                [ $# -gt 1 ] || usage
                ARG_NAME=$(echo "$1" | sed 's/\-\-probe\.//')
                ARG_VALUE="$2"
                shift
            fi
            PROBE_ARGS="$PROBE_ARGS -$ARG_NAME=$ARG_VALUE"
            ;;
        --service-token*)
            if echo "$1" | grep "=" 1>/dev/null; then
                ARG_VALUE=$(echo "$1" | sed 's/\-\-service-token=\(.*\)/\1/')
            else
                [ $# -gt 1 ] || usage
                ARG_VALUE="$2"
                shift
            fi
            PROBE_ARGS="$PROBE_ARGS -token=$ARG_VALUE"
            TOKEN_PROVIDED=true
            touch /etc/service/app/down
            ;;
        *)
            break
            ;;
    esac
    shift
done

echo "$APP_ARGS" >/etc/weave/scope-app.args
echo "$PROBE_ARGS" >/etc/weave/scope-probe.args

# End of the command line can optionally be some
# addresses of apps to connect to, for people not
# using Weave DNS. We stick these in /etc/weave/apps
# for the run-probe script to pick up.
MANUAL_APPS=$@

# Implicitly target the Scope Service if a service token was provided with
# no explicit manual app.
if [ "$MANUAL_APPS" = "" -a "$TOKEN_PROVIDED" = "true" ]; then
    MANUAL_APPS="scope.weave.works:443"
fi

echo "$MANUAL_APPS" >>/etc/weave/apps

exec /home/weave/runsvinit
