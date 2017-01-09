#! /bin/bash

# shellcheck disable=SC1091
. ./config.sh

start_suite "Test some key topologies are not empty"

scope_on "$HOST1" launch

wait_for_containers "$HOST1" 60 weavescope

topology_is_not_empty() {
	local host="$1"
	local topology="$2"
	local timeout="${5:-60}"

	for _ in $(seq "$timeout"); do
		local report
		local count
		report="$(curl -s "http://$host:4040/api/report")"
		count=$(echo "$report" | jq -r ".$topology.nodes | length > 0" 2>/dev/null)

		if [ "$count" = "true" ]; then
      assert "curl -s http://$host:4040/api/report | jq -r '.$topology.nodes | length > 0'" true
			return
		fi
		sleep 1
	done

	echo "Failed to find any nodes in the $topology topology after $timeout secs"
  assert "curl -s http://$host:4040/api/report | jq -r '.$topology.nodes | length > 0'" true
}

topology_is_not_empty "$HOST1" Endpoint
topology_is_not_empty "$HOST1" Process
topology_is_not_empty "$HOST1" Container
topology_is_not_empty "$HOST1" ContainerImage
topology_is_not_empty "$HOST1" Host

scope_end_suite
