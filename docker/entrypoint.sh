#!/bin/sh

usage() {
	echo "$0 --dns <IP> --hostname <NAME> --searchpath <SEARCHPATH> --app.foo bar --probe.foo bar"
	exit 1
}

# This script exists to modify the network settings in the scope containers
# as docker doesn't allow it when started with --net=host

APP_ARGS=""
PROBE_ARGS=""

while true; do
    case "$1" in
        --dns)
            [ $# -gt 1 ] || usage
            DNS_SERVER="$2"
            shift 2
            ;;
        --searchpath)
            [ $# -gt 1 ] || usage
            SEARCHPATH="$2"
            shift 2
            ;;
        --app.*)
            [ $# -gt 1 ] || usage
            ARG_NAME=$(echo "$1" | sed 's/\-\-app\.//')
            ARG_VALUE="$2"
            shift 2
            APP_ARGS="$APP_ARGS -$ARG_NAME=$ARG_VALUE"
            ;;
        --probe.*)
            [ $# -gt 1 ] || usage
            ARG_NAME=$(echo "$1" | sed 's/\-\-probe\.//')
            ARG_VALUE="$2"
            shift 2
            PROBE_ARGS="$PROBE_ARGS -$ARG_NAME=$ARG_VALUE"
            ;;
        --no-app)
            touch /etc/service/app/down
            shift 1
            ;;
        --no-probe)
            touch /etc/service/probe/down
            shift 1
            ;;
        *)
            break
            ;;
    esac
done

mkdir -p /etc/weave
echo "$APP_ARGS" >/etc/weave/scope-app.args
echo "$PROBE_ARGS" >/etc/weave/scope-probe.args

if [ -n "$DNS_SERVER" -a -n "$SEARCHPATH" ]; then
    echo "domain $SEARCHPATH" >/etc/resolv.conf
    echo "search $SEARCHPATH" >>/etc/resolv.conf
    echo "nameserver $DNS_SERVER" >>/etc/resolv.conf
fi

# End of the command line can optionally be some
# addresses of apps to connect to, for people not
# using Weave DNS. We stick these in /etc/weave/apps
# for the run-probe script to pick up.
MANUAL_APPS=$@
echo "$MANUAL_APPS" >/etc/weave/apps

exec /sbin/runsvdir /etc/service
