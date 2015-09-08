package docker

import (
	"strconv"

	"github.com/weaveworks/scope/probe/proc"
	"github.com/weaveworks/scope/probe/process"
	"github.com/weaveworks/scope/report"
)

// Node metadata keys.
const (
	ContainerID = "docker_container_id"
	Domain      = "domain" // TODO this is ambiguous, be more specific
	Name        = "name"   // TODO this is ambiguous, be more specific
)

// These vars are exported for testing.
var (
	NewProcessTreeStub = proc.NewTree
)

// Tagger is a tagger that tags Docker container information to process
// nodes that have a PID.
type Tagger struct {
	registry   Registry
	procWalker proc.ProcReader
}

// NewTagger returns a usable Tagger.
func NewTagger(registry Registry, procWalker proc.ProcReader) *Tagger {
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

func (t *Tagger) tag(tree proc.Tree, topology *report.Topology) {
	for nodeID, nodeMetadata := range topology.Nodes {
		pidStr, ok := nodeMetadata.Metadata[process.PID]
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

		md := report.MakeNodeWith(map[string]string{
			ContainerID: c.ID(),
		})

		topology.Nodes[nodeID] = nodeMetadata.Merge(md)
	}
}
