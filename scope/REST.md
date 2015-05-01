# Rest structure

The various REST calls to retrieve topology information.

# Topologies list

```
GET /api/topology
```

Returns a list of topology descriptions. Each description has the fields:

- `name` - string; friendly name of the topology
- `url` - string; path of the topology
- `grouped_url` - string or not existing; path to the 'grouped' version on the
  topology, if there is one for this topology
- `stats` - object; basic statistics about the topology, with these fields:
    - node_count - int
    - nonpseudo_node_count - int
    - edge_count - int


# Full topologies

```
GET /api/topology/:kind
```

Where `:kind` is one of

- `processpid` - processes by PID (most granular)
- `processname` - processes by name
- `networkip` - hosts (interfaces) by IP
- `networkhost` - hosts by hostname (least granular)

A GET on a topology URL gives:

```json
{
    "nodes": {
        "nodeID": {}
    }
}
```

`nodeID` is an opaque string. The node object is a map with the fields:

- `id` - string; same as nodeID
- `adjacency` - array of strings; list of node IDs that this node is connected to
- `label_major` - string; primary label for this node
- `label_minor` - string; extra label information for this node
- `pseudo` - bool; true if the node is "deduced" from conditions and has no actual probe; false if this node comes from a probe
- `origin` - optional array of strings; origin IDs that the node information comes from

All node IDs listed in `adjacency` have a corresponding node entry. In other
words, the graph is complete.

If A has B in its `adjacency` list, that means A communicates to B. B may or
may not have A in its adjacency list. In other words, the graph is directed.

Pseudonodes are nodes for which we don't have direct measurements, but we know
that other nodes communicate with them. These nodes will always have empty
adjacency and origin lists.


# Websocket

There is also a websocket version of the full topology, which provides
continuous updates.

```
GET ws://hostname/api/topology/:kind/ws?t=5s
```

The URL is the same as for an HTTP topology, but with `/ws` appended. The
argument `t` can be used to specify the update interval.

The topology is send as a stream of updates, starting with the difference
between the empty topology and the current state.

```json
{
    "add": [],
    "update": [],
    "remove": []
}
```

The object has these fields:

- add - array of objects; a list of node objects which are new since the last update
- update - array of objects; a list of node objects which have changed
- remove - array of strings; a list of node IDs which are gone


# Nodes

Given any base topology URL, there is also:

```
GET /api/topology/:kind/:id
```

This returns a single node object. It has fields in addition to those from the
corresponding entry from the full topology.

```json
{
    "node": {}
}
```

Extra fields:
    - aggregate - object - sum of all edge metadata, see 'metadata' below for
      the fields.

# Edges

Details for an edge:

```
GET /api/topology/:kind/:id/:adjacentid
```

This returns the metadata for the undirected edge between :id and :adjacentid,
from the point of view of :id. If the edge is not found it'll still return a
statuscode 200, but without values.

```json
{
    "metadata": {}
}
```

metadata can have these keys:

 - max_conn_count_tcp - int; maximum number of concurrent TCP connections.
 - egress_bytes - int; total number of bytes.
 - ingress_bytes - int; total number of bytes.

Keys will only be set if the statistic is measured by the probes used. If a
statistic has the value 0 the key will still be present.

# Origin

Origin calls give information about the machine on which the measurements are
made.

```
GET /api/origin/:id
```

Origin is a reference to a physical machine that runs a probe. The origin ID
comes from the `origin` field in a node object. The returned object is a map
with the fields:

- `hostname` - string; fully qualified domain name of the host
- `os` - string; operating system of the host, e.g. "linux"
