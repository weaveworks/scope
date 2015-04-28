# Enhance

## 2014 10 05

Informed by a new context, we began sketching out a simplified data model.
Initial requirements:

- Must support order-of-thousands nodes in the topology.
- Must support a scrapable API.
- Packet counts, etc. are a value-add and should be composed in to the topology as needed.

Desires based on experience so far:

- Too much effort was spent wrangling deeply nested hierarchies in the report structs. They should be flattened.
- It would be nice if the probe API were directly consumable by the app.
  - This enables much simpler composition of "bridge" components, including none at all.
  - This composes nicely with the scrapable API requirement.
- The simplest use case of no metadata at all could (should) be satisfied by scraping the /proc FS, with no packet sniffing.

Peter's initial thoughts was to key everything based on the atom of host + socket_descriptor.
This represents the most specific (application-level) topology.
Other more general topologies (network, transport) could be derived from that.

Harmen noted that we want to capture data that may not be associated with a socket_descriptor.
Therefore we should build multiple, independent topologies: application, network, and transport.
Each one is a complete topology for that layer, with lower (OSI) layers always containing the higher ones.

We think we can get away with a report structure that looks like this

```json
{
	"application": {
		"adjacency": {
			"host:src": ["dst1", "dst2", "dst3"]
		},
		"metadata": {
			"host:src:dst1": {}
		}
	}

	"network": {
		"adjacency": {},
		"metadata": {}
	}

	"transport": {
		"adjacency": {},
		"metadata": {}
	}
}
```

Where the meaning of the keys and key-parts depends on the section of the report.
Specifically, in the application section, src is an encoding of the socket_descriptor.

Note that we still need another section to map application keys to processes, either physical (PIDs) or logical (process names, service names, etc.).
That mapping can take the form of a function:

```go
type mapper func(socket_descriptor) string
```

Where the returned string will be the aggregation key of the logical (application) topology.
At the lowest level, it could be the PID that bound the socket_descriptor.
At a middle layer, it could be the name of the process associated with the PID.
At a higher layer, the function could query a service discovery component, and return a service name.
It would probably be good to support multiple simultaneous mappers, so the application-layer topology could have multiple levels of detail.

### Example implementation

Two probes, A and B, with a single TCP connection, both agree there is a connection.

```json
{
	"proc": {
        "mapping": {
            "hostA:192.168.1.1:12345": "curl",
            "hostB:192.168.1.2:80": "apache"
        },
		"adjacency": {
            "hostA:192.168.1.1:12345": ["192.168.1.2:80"],
            "hostB:192.168.1.2:80": ["192.168.1.1:12345"]
		},
		"metadata": {
            "hostA:192.168.1.1:12345:192.168.1.2:80": {"egress": "12bytes", "ingress": "0", "d": "15s"},
            "hostB:192.168.1.2:80:192.168.1.1:12345": {"egress": "0", "ingress": "12bytes", "d": "15s"}
        }
	},

	"transport": {
		"adjacency": {
            "hostA:192.168.1.1:12345": ["192.168.1.2:80"],
            "hostB:192.168.1.2:80": ["192.168.1.1:12345"]
        },
		"metadata": {
            "hostA:192.168.1.1:12345:192.168.1.2:80": {"egress": "12bytes", "ingress": "0", "d": "15s"},
            "hostB:192.168.1.2:80:192.168.1.1:12345": {"egress": "0", "ingress": "12bytes", "d": "15s"}
        }
	},

	"network": {
		"adjacency": {
            "hostA:192.168.1.1": ["192.168.1.2"],
            "hostB:192.168.1.2": ["192.168.1.1"]
        },
		"metadata": {
            "hostA:192.168.1.1:12345:192.168.1.2:80": {"egress": "12bytes", "ingress": "0", "d": "15s"},
            "hostB:192.168.1.2:80:192.168.1.1:12345": {"egress": "0", "ingress": "12bytes", "d": "15s"}
        }
	},

    "nodes": {
        "hostA": {
            "addresses": ["192.168.1.1"],
            "hostname": "..."
        },
        "hostB": {
            "addresses": ["192.168.1.2"],
            "hostname": "...",
        },
    }
}
```

Note that the proc.mapping can only be filled by the procfs-probe.
But {proc,transport}.{adjacency,metadata} could in theory be filled both procfs-probe and sniffing-probe.
That's a consequence of the fact that, in the case of UDP and TCP, the socket_descriptor is fully reconstructable from a network packet's source and destination fields.
For simplicity, and for now, we'd like to restrict that freedom, and state that the proc section should be filled only by the procfs-probe.
This draws a clear line of ownership and demarcation.
We can revisit this decision in the future, if it proves too restrictive.

For now we will assert that the metadata type is the same structure for each section.
Adjacency list values don't necessarily need to exist as nodes. Cases are non-probe machines and broadcast addresses.

Merging reports erquires a time component because rates require a window, and we can't make assumptions about fixed window sizes (either because the edge isn't present for the entire interval, or the whole probe isn't).
Reports are by definition non-overlapping for a single e.g. socket_descriptor, so that's a safe assumption to make. Therefore, all we need are durations, which may be semantically merged with simple addition.

