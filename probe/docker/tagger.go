package docker

import (
	"strconv"

	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// These constants are keys used in node metadata
// TODO: use these constants in report/{mapping.go, detailed_node.go} - pending some circular references
const (
	ContainerID = "docker_container_id"
)

// These vars are exported for testing.
var (
	NewProcessTreeStub = process.NewTree
)

// Tagger is a tagger that tags Docker container information to process
// nodes that have a PID.
type Tagger struct {
	registry   Registry
	procWalker process.Walker
}

// NewTagger returns a usable Tagger.
func NewTagger(registry Registry, procWalker process.Walker) *Tagger {
	return &Tagger{
		registry:   registry,
		procWalker: procWalker,
	}
}

// Tag implements Tagger.
func (t *Tagger) Tag(r report.Report) (report.Report, error) {
	tree, err := NewProcessTreeStub(t.procWalker)
	if err != nil {
		return report.MakeReport(), err
	}
	t.tag(tree, &r.Process)
	return r, nil
}

func (t *Tagger) tag(tree process.Tree, topology *report.Topology) {
	for nodeID, nodeMetadata := range topology.NodeMetadatas {
		pidStr, ok := nodeMetadata["pid"]
		if !ok {
			continue
		}

		pid, err := strconv.ParseUint(pidStr, 10, 64)
		if err != nil {
			continue
		}

		var (
			c         Container
			candidate = int(pid)
		)

		t.registry.LockedPIDLookup(func(lookup func(int) Container) {
			for {
				c = lookup(candidate)
				if c != nil {
					break
				}

				candidate, err = tree.GetParent(candidate)
				if err != nil {
					break
				}
			}
		})

		if c == nil {
			continue
		}

		md := report.NodeMetadata{
			ContainerID: c.ID(),
		}

		topology.NodeMetadatas[nodeID].Merge(md)
	}
}
