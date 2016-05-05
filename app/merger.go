package app

import (
	"fmt"
	"math"
	"sort"

	"github.com/bluele/gcache"
	"github.com/spaolacci/murmur3"

	"github.com/weaveworks/scope/report"
)

// Merger is the type for a thing that can merge reports.
type Merger interface {
	Merge([]report.Report) report.Report
}

type dumbMerger struct{}

// MakeDumbMerger makes a Merger which merges together reports in the simplest possible way.
func MakeDumbMerger() Merger {
	return dumbMerger{}
}

func (dumbMerger) Merge(reports []report.Report) report.Report {
	rpt := report.MakeReport()
	id := murmur3.New64()
	for _, r := range reports {
		rpt = rpt.Merge(r)
		id.Write([]byte(r.ID))
	}
	rpt.ID = fmt.Sprintf("%x", id.Sum64())
	return rpt
}

type smartMerger struct {
	cache gcache.Cache
}

// NewSmartMerger makes a Merger which merges together reports, caching intermediate merges
// to accelerate future merges. Idea is to cache pair-wise merged reports, forming a merge
// tree.  Merging a new report into this tree should be log(N).
func NewSmartMerger() Merger {
	return &smartMerger{
		cache: gcache.New(1000).LRU().Build(),
	}
}

type node struct {
	id  uint64
	rpt report.Report
}

type byID []*node

func (ns byID) Len() int           { return len(ns) }
func (ns byID) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns byID) Less(i, j int) bool { return ns[i].id < ns[j].id }

func hash(ids ...string) uint64 {
	id := murmur3.New64()
	for _, i := range ids {
		id.Write([]byte(i))
	}
	return id.Sum64()
}

func (s *smartMerger) ClearCache() {
	s.cache.Purge()
}

func (s *smartMerger) Merge(reports []report.Report) report.Report {
	// Start with a sorted list of leaves.
	// Note we must dedupe reports with the same ID to ensure the
	// algorithm below doesn't go into an infinite loop.  This is
	// fine as reports with the same ID are assumed to be the same.
	nodes := []*node{}
	seen := map[uint64]struct{}{}
	for _, r := range reports {
		id := hash(r.ID)
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		nodes = append(nodes, &node{
			id:  id,
			rpt: r,
		})
	}
	sort.Sort(byID(nodes))

	// Define how to merge two nodes together.  The result of merging
	// two reports is cached.
	merge := func(left, right *node) *node {
		id := hash(left.rpt.ID, right.rpt.ID)

		if result, err := s.cache.Get(id); err == nil {
			return result.(*node)
		}

		n := &node{
			id:  id,
			rpt: report.MakeReport().Merge(left.rpt).Merge(right.rpt),
		}
		s.cache.Set(id, n)
		return n
	}

	// Define how to reduce n nodes to 1.
	// Min and max are both inclusive!
	var reduce func(min, max uint64, nodes []*node) *node
	reduce = func(min, max uint64, nodes []*node) *node {
		switch len(nodes) {
		case 0:
			return &node{rpt: report.MakeReport()}
		case 1:
			return nodes[0]
		case 2:
			return merge(nodes[0], nodes[1])
		}

		partition := min + ((max - min) / 2)
		index := sort.Search(len(nodes), func(i int) bool {
			return nodes[i].id > partition
		})
		if index == len(nodes) {
			return reduce(min, partition, nodes)
		} else if index == 0 {
			return reduce(partition+1, max, nodes)
		}
		left := reduce(min, partition, nodes[:index])
		right := reduce(partition+1, max, nodes[index:])
		return merge(left, right)
	}

	return reduce(0, math.MaxUint64, nodes).rpt
}
