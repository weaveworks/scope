#!/bin/bash

set -eu

if [ $# -ne 1 ]; then
	echo "Usage: $0 <host>"
	exit 1
fi

HOST=$1

echo "Starting proxy container..."
PROXY_CONTAINER=$(ssh "$HOST" weave run -d weaveworks/socksproxy)

function finish {
	echo "Removing proxy container.."
	# shellcheck disable=SC2029
	ssh "$HOST" docker rm -f "$PROXY_CONTAINER"
}
trap finish EXIT

# shellcheck disable=SC2029
PROXY_IP=$(ssh "$HOST" -- "docker inspect --format='{{.NetworkSettings.IPAddress}}' $PROXY_CONTAINER")
echo 'Please configure your browser for proxy http://localhost:8080/proxy.pac'
# shellcheck disable=SC2029
ssh "-L8000:$PROXY_IP:8000" "-L8080:$PROXY_IP:8080" "$HOST" docker attach "$PROXY_CONTAINER"
