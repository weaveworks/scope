package report

import "strconv"

// MakeDetailedNode transforms a renderable node to a detailed node. It uses
// aggregate metadata, plus the set of origin node IDs, to produce tables.
func MakeDetailedNode(r Report, n RenderableNode) DetailedNode {
	tables := []Table{}
	{
		rows := []Row{}
		if val, ok := n.Metadata[KeyMaxConnCountTCP]; ok {
			rows = append(rows, Row{"TCP connections", strconv.FormatInt(int64(val), 10), ""})
		}
		if val, ok := n.Metadata[KeyBytesIngress]; ok {
			rows = append(rows, Row{"Bytes ingress", strconv.FormatInt(int64(val), 10), ""})
		}
		if val, ok := n.Metadata[KeyBytesEgress]; ok {
			rows = append(rows, Row{"Bytes egress", strconv.FormatInt(int64(val), 10), ""})
		}
		if len(rows) > 0 {
			tables = append(tables, Table{"Connections", true, rows})
		}
	}

	// RenderableNode may be the result of merge operation(s), and so may have
	// multiple origins. The ultimate goal here is to generate tables to view
	// in the UI, so we skip the intermediate representations, but we could
	// add them later.
	for _, id := range n.Origins {
		if table, ok := OriginTable(r, id); ok {
			// Origin node IDs are unique, so we'll be optimistic, here, and
			// assume they'll also produce unique tables.
			tables = append(tables, table)
		}
	}

	return DetailedNode{
		ID:         n.ID,
		LabelMajor: n.LabelMajor,
		LabelMinor: n.LabelMinor,
		Pseudo:     n.Pseudo,
		Tables:     tables,
	}
}

// OriginTable produces a table (to be consumed directly by the UI) based on
// an origin ID, which is (optimistically) a node ID in one of our topologies.
func OriginTable(r Report, originID string) (Table, bool) {
	if nmd, ok := r.Endpoint.NodeMetadatas[originID]; ok {
		return endpointOriginTable(nmd)
	}
	if nmd, ok := r.Address.NodeMetadatas[originID]; ok {
		return addressOriginTable(nmd)
	}
	if nmd, ok := r.Process.NodeMetadatas[originID]; ok {
		return processOriginTable(nmd)
	}
	if nmd, ok := r.Container.NodeMetadatas[originID]; ok {
		return containerOriginTable(nmd)
	}
	if nmd, ok := r.Host.NodeMetadatas[originID]; ok {
		return hostOriginTable(nmd)
	}
	return Table{}, false
}

func endpointOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{"endpoint", "Endpoint"},
		{"host_name", "Host name"},
		{"pid", "PID"},
		{"name", "Process name"},
	} {
		if val, ok := nmd[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}
	return Table{
		Title:   "Origin Endpoint",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func addressOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["address"]; ok {
		rows = append(rows, Row{"Address", val, ""})
	}
	if val, ok := nmd["host_name"]; ok {
		rows = append(rows, Row{"Host name", val, ""})
	}
	return Table{
		Title:   "Origin Address",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func processOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["comm"]; ok {
		rows = append(rows, Row{"Name (comm)", val, ""})
	}
	if val, ok := nmd["pid"]; ok {
		rows = append(rows, Row{"PID", val, ""})
	}
	if val, ok := nmd["ppid"]; ok {
		rows = append(rows, Row{"Parent PID", val, ""})
	}
	return Table{
		Title:   "Origin Process",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func containerOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	for _, tuple := range []struct{ key, human string }{
		{"docker_container_id", "Container ID"},
		{"docker_container_name", "Container name"},
		{"docker_image_id", "Container image ID"},
		{"docker_image_name", "Container image name"},
	} {
		if val, ok := nmd[tuple.key]; ok {
			rows = append(rows, Row{Key: tuple.human, ValueMajor: val, ValueMinor: ""})
		}
	}
	return Table{
		Title:   "Origin Container",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}

func hostOriginTable(nmd NodeMetadata) (Table, bool) {
	rows := []Row{}
	if val, ok := nmd["host_name"]; ok {
		rows = append(rows, Row{"Host name", val, ""})
	}
	if val, ok := nmd["load"]; ok {
		rows = append(rows, Row{"Load", val, ""})
	}
	if val, ok := nmd["os"]; ok {
		rows = append(rows, Row{"Operating system", val, ""})
	}
	return Table{
		Title:   "Origin Host",
		Numeric: false,
		Rows:    rows,
	}, len(rows) > 0
}
