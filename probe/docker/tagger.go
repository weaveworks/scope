package docker

import (
	"strconv"

	"$GITHUB_URI/probe/process"
	"$GITHUB_URI/report"
)

// Node metadata keys.
const (
	ContainerID = "docker_container_id"
	Domain      = "domain" // TODO this is ambiguous, be more specific
	Name        = "name"   // TODO this is ambiguous, be more specific
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

// Name of this tagger, for metrics gathering
func (Tagger) Name() string { return "Docker" }

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
	for nodeID, node := range topology.Nodes {
		pidStr, ok := node.Latest.Lookup(process.PID)
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

		if c == nil || ContainerIsStopped(c) || c.PID() == 1 {
			continue
		}

		topology.AddNode(report.MakeNodeWith(nodeID, map[string]string{
			ContainerID: c.ID(),
		}).WithParents(report.EmptySets.
			Add(report.Container, report.MakeStringSet(report.MakeContainerNodeID(c.ID()))).
			Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(c.Image()))),
		))

	}
}
