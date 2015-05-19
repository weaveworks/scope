#!/bin/sh

usage() {
	echo "$0 --dns <IP> --hostname <NAME> --searchpath <SEARCHPATH>"
	exit 1
}

# This script exists to modify the network settings in the scope containers
# as docker doesn't allow it when started with --net=host
while true; do
    case "$1" in
        --dns)
            [ $# -gt 1 ] || usage
            DNS_SERVER="$2"
            shift 2
            ;;
        --hostname)
            [ $# -gt 1 ] || usage
            HOSTNAME="$2"
            shift 2
            ;;
        --searchpath)
            [ $# -gt 1 ] || usage
            SEARCHPATH="$2"
            shift 2
            ;;
        *)
            break
            ;;
    esac
done

if [ -n "$DNS_SERVER" -a -n "$SEARCHPATH" ]; then
    echo "domain $SEARCHPATH" >/etc/resolv.conf
    echo "search $SEARCHPATH" >>/etc/resolv.conf
    echo "nameserver $DNS_SERVER" >>/etc/resolv.conf
fi

if [ -n "$HOSTNAME" ]; then
    echo "$HOSTNAME" >/etc/hostname
    hostname -F /etc/hostname
fi

exec /sbin/runsvdir /etc/service
