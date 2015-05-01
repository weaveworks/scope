# Cello concepts and design

Cello is comprised of three parts.

- **cello-app** acts a single interface over multiple cello-bridge instances,
  providing a unified view of the complete network [topology](#topology).
- **cello-bridge** collects reports from cello-probes and provides a queryable
  API to the cello-app
- **cello-probe** sniffs traffic on each host in your network and publishes
  regular reports to cello-bridge.

## Topology

The topology is a collection of endpoints, hosts, and classes, and the edges
between each.

### Endpoint

The lowest-level node, representing a set of network addresses. Usually
exactly one, but can be many when an endpoint represents a whole subnetwork.
An endpoint has a unique ID and miscellaneous metadata, like addresses. An
endpoint may belong to a host, but it's not required.

### Host

The mid-level node, representing a physical machine. A host has a unique ID,
miscellaneous metadata, and zero or more endpoints (though a host with zero
endpoints would be a curious thing indeed). A host belongs to zero or more
classes.

### Class

The top-level node, representing a collection of hosts that share certain
properties. A class has a unique ID, miscellaneous metadata, and zero or more
hosts (though a class with zero hosts would be a curious thing indeed).

Classes are derived from classification rules, or classifiers.

### Endpoint-to-endpoint edges

The lowest-level edge, and the edges that are reported by the probe processes,
endpoints-to-endpoints (E2E) edges are non-directional links between two
endpoints, and all of the raw, first-order metrics about the traffic visible on
that link.

All edges have the structure:

- ID
- A (from the perspective of node A)
-- Ingress report
-- Egress report
- B (from the perspective of node B)
-- Ingress report
-- Egress report

### Host-to-host edges

The mid-level edge, Host-to-host (H2H) edges are non-directional edges derived
by `app` aggregating all E2E edges from a specific pair of hosts.

### Class-to-class edges

The top-level edge, class-to-class (C2C) edges are non-directional edges derived
by `app` aggregating all H2H edges from all hosts that belong in one class to
all hosts that belong to another class.

## Transforms

There are two type of transforms, but both are based on the concept of a
[rule](#rule): classifiers and filters. [Classifiers](#classifier) take a
topology without classes, and add the classes based on rules.
[Filters](#filter) take a topology with classes, as returned by the
classifiers, and remove edges, according to rules. Both classifiers and
filters us the same rule concept, but are configured with a different set of
rules.

### Rule

Given an abstract definition of an edge, satisfiable by all of E2E, H2H, and
C2C edges, a **rule** is a function that takes an edge and zero, one, or two
of its connected nodes.

For more on rules, see the [rules document][rules].

[rules]: ./rules.html

### Filter

Filters are rules that are used to remove nodes from the topology. If a node
passes a filter rule, it's either shown exclusively, or hidden, depending if
the filter is filter-in or filter-out.

### Classifier

Classifiers are rules that are used to group nodes. If a node passes a
classifier rule, it's _tagged_ by that classifier, and becomes a member of
that classifier's group. A single node can be a member of zero or more
classifier groups.


## Note on edges

In order to merge edges, it's necessary that the data associated with an edge
is mergable with other edge-data. Concretely, that means edge reports should
contain only summable or countable data (e.g. byte counts), and no derived
data (e.g. throughput rates). That also implies information to convert raw
edge-wise data to derived level-wise data, e.g. window over which the data was
collected, should be passed with the edges, as opposed to embedded within
them.
