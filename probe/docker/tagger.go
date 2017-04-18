package docker

import (
	"strconv"
	"strings"

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
	NewProcessTreeStub = process.NewTree
)

// Tagger is a tagger that tags Docker container information to process
// nodes that have a PID.
// It also populates the SwarmService topology if any of the associated docker labels are present.
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

	// Scan for Swarm service info
	for containerID, container := range r.Container.Nodes {
		serviceID, ok := container.Latest.Lookup(LabelPrefix + "com.docker.swarm.service.id")
		if !ok {
			continue
		}
		serviceName, ok := container.Latest.Lookup(LabelPrefix + "com.docker.swarm.service.name")
		if !ok {
			continue
		}
		stackNamespace, ok := container.Latest.Lookup(LabelPrefix + "com.docker.stack.namespace")
		if !ok {
			continue
		}

		prefix := stackNamespace + "_"
		if strings.HasPrefix(serviceName, prefix) {
			serviceName = serviceName[len(prefix):]
		}

		nodeID := report.MakeSwarmServiceNodeID(serviceID)
		node := report.MakeNodeWith(nodeID, map[string]string{
			ServiceName:    serviceName,
			StackNamespace: stackNamespace,
		})
		r.SwarmService = r.SwarmService.AddNode(node)

		r.Container.Nodes[containerID] = container.WithParents(container.Parents.Add(report.SwarmService, report.MakeStringSet(nodeID)))
	}

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

		node := report.MakeNodeWith(nodeID, map[string]string{
			ContainerID: c.ID(),
		}).WithParents(report.EmptySets.
			Add(report.Container, report.MakeStringSet(report.MakeContainerNodeID(c.ID()))),
		)

		// If we can work out the image name, add a parent tag for it
		image, ok := t.registry.GetContainerImage(c.Image())
		if ok && len(image.RepoTags) > 0 {
			imageName := ImageNameWithoutVersion(image.RepoTags[0])
			node = node.WithParents(report.EmptySets.
				Add(report.ContainerImage, report.MakeStringSet(report.MakeContainerImageNodeID(imageName))),
			)
		}

		topology.AddNode(node)
	}
}
