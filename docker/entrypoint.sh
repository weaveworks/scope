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
        *)
            break
            ;;
    esac
done

mkdir -p /etc/weave
echo "$APP_ARGS" >/etc/weave/app.args
echo "$PROBE_ARGS" >/etc/weave/probe.args

if [ -n "$DNS_SERVER" -a -n "$SEARCHPATH" ]; then
    echo "domain $SEARCHPATH" >/etc/resolv.conf
    echo "search $SEARCHPATH" >>/etc/resolv.conf
    echo "nameserver $DNS_SERVER" >>/etc/resolv.conf
fi

# End of the command line can optionally be some
# addresses of probes to connect to, for people not
# using Weave DNS.  We stick these in /etc/weave/probes
# for the run-app script to pick up.
MANUAL_PROBES=$@
echo "$MANUAL_PROBES" >/etc/weave/probes

exec /sbin/runsvdir /etc/service
