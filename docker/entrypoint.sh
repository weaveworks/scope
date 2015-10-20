#!/bin/sh

usage() {
	echo "$0 --app.foo bar --probe.foo bar"
	exit 1
}

# This script exists to modify the network settings in the scope containers
# as docker doesn't allow it when started with --net=host

WEAVE_CONTAINER_NAME=weave
DOCKER_BRIDGE=docker0
HOSTNAME=scope
DOMAIN=weave.local
IP_REGEXP="[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}"

container_ip() {
    if ! status=$(docker inspect --format='{{.State.Running}} {{.HostConfig.NetworkMode}}' $1 2>/dev/null); then
        echo "Container $1 not found" >&2
        return 1
    fi
    case "$status" in
        "true host")
            CONTAINER_IP="127.0.0.1"
            return 0
            ;;
        "true default" | "true bridge")
            CONTAINER_IP="$(docker inspect --format='{{.NetworkSettings.IPAddress}}' $1 2>/dev/null)"
            return 0
            ;;
        *)
            echo "Container $1 not running" >&2
            return 1
            ;;
    esac
}

is_running() {
    status=$(docker inspect --format='{{.State.Running}}' $1 2>/dev/null) && [ "$status" = "true" ]
    return $?
}

docker_bridge_ip() {
    local DOCKER_BRIDGE_IP=$(ip -f inet address show dev $DOCKER_BRIDGE | grep -m1 -o 'inet \([.0-9]\)*')
    echo ${DOCKER_BRIDGE_IP#inet }
}

# Run `weave` in the weave exec container
weave() {
    WEAVEXEC_IMAGE=$(docker inspect --format='{{.Config.Image}}' weave | sed 's/\/weave/\/weaveexec/')
    docker run -t --rm --privileged --net=host \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v /proc:/hostproc \
        -e PROCFS=/hostproc \
        $WEAVEXEC_IMAGE --local "$@"
}

# Run `weave expose` if it's not already exposed.
weave_expose() {
    status=$(weave ps weave:expose | awk '{print $3}' 2>/dev/null)
    if [ "$status" = "" ]; then
        echo "Exposing host to weave network."
        weave expose
    fi
}

mkdir -p /etc/weave
APP_ARGS=""
PROBE_ARGS=""

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
            echo "scope.weave.works:443" >/etc/weave/apps
            touch /etc/service/app/down
            ;;
        --no-app)
            touch /etc/service/app/down
            ;;
        --no-probe)
            touch /etc/service/probe/down
            ;;
        *)
            break
            ;;
    esac
    shift
done

if is_running $WEAVE_CONTAINER_NAME; then
    container_ip $WEAVE_CONTAINER_NAME
    PROBE_ARGS="$PROBE_ARGS -weave.router.addr=$CONTAINER_IP"
    weave_expose

    DOCKER_BRIDGE_IP=$(docker_bridge_ip)
    echo "Weave container detected at $CONTAINER_IP, Docker bridge at $DOCKER_BRIDGE_IP"

    echo "domain $DOMAIN" >/etc/resolv.conf
    echo "search $DOMAIN" >>/etc/resolv.conf
    echo "nameserver $DOCKER_BRIDGE_IP" >>/etc/resolv.conf

    IP_ADDRS=$(find /sys/class/net -type l | xargs -n1 basename | grep -vE 'docker|veth|lo' | \
        xargs -n1 ip addr show | grep inet | awk '{ print $2 }' | grep -oE "$IP_REGEXP")
    CONTAINER=$(docker inspect --format='{{.Id}}' weavescope)
    if [ -z "$IP_ADDRS" ]; then
        echo "Could not determine local IP address; Weave DNS integration will not work correctly."
        exit 1
    else
        for ip in $IP_ADDRS; do
            weave dns-add $ip $CONTAINER -h $HOSTNAME.$DOMAIN
        done
    fi
fi

echo "$APP_ARGS" >/etc/weave/scope-app.args
echo "$PROBE_ARGS" >/etc/weave/scope-probe.args

# End of the command line can optionally be some
# addresses of apps to connect to, for people not
# using Weave DNS. We stick these in /etc/weave/apps
# for the run-probe script to pick up.
MANUAL_APPS=$@
echo "$MANUAL_APPS" >>/etc/weave/apps

exec /home/weave/runsvinit

