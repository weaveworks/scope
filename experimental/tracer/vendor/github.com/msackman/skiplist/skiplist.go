package skiplist

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

const (
	p            = 0.3
	defaultDepth = 2
)

type Comparable interface {
	LessThan(Comparable) bool
	Equal(Comparable) bool
}

type SkipList struct {
	length             uint
	terminus           *Node
	levelProbabilities []float32
	curCapacity        uint
	curDepth           uint
	nodes              []Node
	localRand          *rand.Rand
}

type Node struct {
	Key        Comparable
	Value      interface{}
	heightRand float32
	prev       *Node
	nexts      []*Node
	skiplist   *SkipList
}

func New(rng *rand.Rand) *SkipList {
	depth := defaultDepth

	terminus := &Node{
		heightRand: 0,
		nexts:      make([]*Node, depth),
	}
	terminus.prev = terminus
	for idx := 0; idx < len(terminus.nexts); idx++ {
		terminus.nexts[idx] = terminus
	}
	s := &SkipList{
		length:    0,
		terminus:  terminus,
		curDepth:  uint(depth),
		localRand: rng,
	}
	s.levelProbabilities = []float32{p}
	terminus.skiplist = s
	s.determineCapacity()

	// s.validate()

	return s
}

func (s *SkipList) determineCapacity() {
	base := float64(1.0) / p
	capacity := math.Pow(base, float64(s.curDepth))
	s.curCapacity = uint(math.Floor(capacity))
}

func (s *SkipList) chooseNumLevels() (float32, int) {
	r := s.localRand.Float32()
	max := len(s.levelProbabilities)
	for idx := 0; idx < max; idx++ {
		if r > s.levelProbabilities[idx] {
			return r, idx + 1
		}
	}
	return r, max + 1
}

func (s *SkipList) ensureCapacity() {
	// defer s.validate()
	if s.length < s.curCapacity {
		return
	}

	threshold := p * s.levelProbabilities[s.curDepth-2]
	s.curDepth++
	s.levelProbabilities = append(s.levelProbabilities, threshold)

	s.determineCapacity()

	// cur and next are just used to walk through the list at lvl. prev
	// records the last node that made it up to the new level.
	cur := s.terminus
	lvl := len(cur.nexts) - 1
	prev := cur
	for {
		next := cur.nexts[lvl]
		if cur.heightRand <= threshold {
			cur.nexts = append(cur.nexts, s.terminus)
			prev.nexts[lvl+1] = cur
			prev = cur
		}
		if next == s.terminus {
			break
		} else {
			cur = next
		}
	}
}

func (s *SkipList) getNode() *Node {
	l := len(s.nodes)
	if l == 0 {
		l = int(s.curCapacity)
		s.nodes = make([]Node, l)
	}
	l--
	n := &s.nodes[l]
	s.nodes = s.nodes[:l]
	return n
}

func (s *SkipList) getEqOrLessThan(cur *Node, k Comparable, captureDescent bool) (*Node, []*Node) {
	// defer s.validate()

	if s.length == 0 {
		return s.terminus, nil
	}
	if cur != s.terminus {
		if k.Equal(cur.Key) {
			return cur, nil
		}
		if k.LessThan(cur.Key) {
			return s.getEqOrLessThan(s.terminus, k, captureDescent)
		}
	}
	// 1. Travel, not-descending, as far as possible
	lvl := len(cur.nexts) - 1
	for {
		n := cur.nexts[lvl]
		if n == s.terminus {
			break
		}
		if n.Key.LessThan(k) {
			cur = n
			lvl = len(cur.nexts) - 1
		} else if n.Key.Equal(k) {
			return n, nil
		} else {
			break
		}
	}
	// 2. Now descend as needed
	var descent []*Node
	if captureDescent {
		descent = make([]*Node, lvl+1)
		descent[lvl] = cur
	}
	for lvl--; lvl >= 0; lvl-- {
		for {
			n := cur.nexts[lvl]
			if n == s.terminus {
				break
			}
			if n.Key.LessThan(k) {
				cur = n
			} else if n.Key.Equal(k) {
				return n, descent
			} else {
				break
			}
		}
		if captureDescent {
			descent[lvl] = cur
		}
	}
	return cur, descent
}

func (s *SkipList) insert(cur *Node, k Comparable, v interface{}, n *Node) *Node {
	// defer s.validate()

	// do this first even though we may not need to - if we do it after
	// the getEqOrLessThan call, we may break descent.
	s.ensureCapacity()
	cur, descent := s.getEqOrLessThan(cur, k, true)
	if cur != s.terminus && cur.Key.Equal(k) {
		cur.Value = v
		return cur
	}
	// We didn't find k, so cur will be the node immediately prior to
	// where k should go.
	heightRand, height := s.chooseNumLevels()
	if n == nil {
		n = s.getNode()
	}
	n.Key = k
	n.Value = v
	n.heightRand = heightRand
	n.nexts = make([]*Node, height)
	n.prev = cur
	n.skiplist = s

	if len(cur.nexts) >= len(n.nexts) {
		for idx := 0; idx < len(n.nexts); idx++ {
			n.nexts[idx] = cur.nexts[idx]
			cur.nexts[idx] = n
		}
	} else {
		// Descent may capture only part of the path: it may be shorter
		// than levels (in the case where the original cur is !=
		// s.terminus) and we reached the correct location without
		// travelling up very far. However, because we didn't find k, we
		// know that all the "lower" levels of descent will be populated
		// (where "lower" is "closer to [0]"), so we just need to fill in
		// the "top".
		if len(n.nexts) > len(descent) {
			_, extra := s.getEqOrLessThan(s.terminus, descent[len(descent)-1].Key, true)
			// Aside: because we know we'll find that Key, all the lower
			// indices of extra will be nil.
			descent = append(descent, extra[len(descent):]...)
		}
		for idx := 0; idx < len(n.nexts); idx++ {
			n.nexts[idx] = descent[idx].nexts[idx]
			descent[idx].nexts[idx] = n
		}
	}
	n._next().prev = n
	s.length++
	return n
}

func (s *SkipList) remove(cur *Node, k Comparable) interface{} {
	// defer s.validate()

	n, _ := s.getEqOrLessThan(cur, k, false)
	if n == s.terminus || !n.Key.Equal(k) {
		return nil
	}
	s.removeNode(n)
	n.nullify()
	return n.Value
}

func (s *SkipList) removeNode(n *Node) {
	// defer s.validate()

	p := n.prev
	n._next().prev = p
	s.length--
	for idx := 0; idx < len(p.nexts) && idx < len(n.nexts); idx++ {
		p.nexts[idx] = n.nexts[idx]
	}
	if len(p.nexts) < len(n.nexts) {
		_, descent := s.getEqOrLessThan(s.terminus, p.Key, true)
		// because we know we're going to find Key, the lower indices
		// of descent will be nil. But we know p == n.prev, so all of
		// those pointers will be to n anyway, which we've already
		// dealt with in the previous loop.
		for idx := len(p.nexts); idx < len(n.nexts); idx++ {
			descent[idx].nexts[idx] = n.nexts[idx]
		}
	}
}

func (s *SkipList) reposition(cur *Node, k Comparable) {
	// defer s.validate()

	needsMove := false
	if cur != s.terminus {
		if cur.prev != s.terminus && !cur.prev.Key.LessThan(k) {
			needsMove = true
		} else if n := cur._next(); n != s.terminus && !k.LessThan(n.Key) {
			needsMove = true
		}
	}
	if needsMove {
		s.removeNode(cur)
		cur.Key = k
		s.insert(cur.prev, cur.Key, cur.Value, cur)
	}
}

func (s *SkipList) First() *Node {
	return s.terminus.Next()
}

func (s *SkipList) Last() *Node {
	return s.terminus.Prev()
}

func (s *SkipList) Insert(k Comparable, v interface{}) *Node {
	return s.insert(s.terminus, k, v, nil)
}

func (s *SkipList) Get(k Comparable) *Node {
	return s.terminus.Get(k)
}

func (s *SkipList) Remove(k Comparable) interface{} {
	return s.remove(s.terminus, k)
}

func (s *SkipList) Len() uint {
	return s.length
}

// NB: this destroys t. Do not use t after this.
func (s *SkipList) Merge(t *SkipList) {
	// defer s.validate()

	cur := s.terminus
	for n := t.First(); n != nil; {
		m := n.Next() // need to save this out before we destroy it in the insert
		cur = s.insert(cur, n.Key, n.Value, n)
		n = m
	}
}

func (s *SkipList) validate() {
	visited := make(map[*Node]bool, int(s.length))
	cur := s.terminus
	visited[cur] = true
	l := uint(0)
	for {
		if cur != s.terminus {
			l++
		}
		if cur._next().prev != cur {
			panic(fmt.Sprintf("Node (%v) has next pointer to %v, which has prev pointer to %v", cur, cur._next(), cur._next().prev))
		}
		if cur.prev._next() != cur {
			panic(fmt.Sprintf("Node (%v) has prev pointer to %v, which has next pointer to %v", cur, cur.prev, cur.prev._next()))
		}
		for h, n := range cur.nexts {
			if h >= len(n.nexts) {
				panic(fmt.Sprintf("Node (%v) has next pointer at level %v pointing down to node (%v) which has %v height", cur, h, n, len(n.nexts)))
			}
		}
		n := cur._next()
		if n == s.terminus {
			break
		}
		if visited[n] {
			panic(fmt.Sprintf("Node (%v) has next as %v which is already visited!", cur, n))
		}
		if cur != s.terminus && !cur.Key.LessThan(n.Key) {
			panic(fmt.Sprintf("Node keys in wrong order: expecting %v < %v", cur.Key, n.Key))
		}
		if n.prev != cur {
			panic(fmt.Sprintf("Node (%v) has next (%v) which does not point back correctly", cur, n))
		}
		cur = n
	}
	if l != s.length {
		panic(fmt.Sprintf("length is wrong: counted %v but length is %v", l, s.length))
	}
}

func (n *Node) Get(k Comparable) *Node {
	m, _ := n.skiplist.getEqOrLessThan(n, k, false)
	if m != n.skiplist.terminus && m.Key.Equal(k) {
		return m
	} else {
		return nil
	}
}

func (n *Node) Insert(k Comparable, v interface{}) *Node {
	return n.skiplist.insert(n, k, v, nil)
}

func (n *Node) Remove() interface{} {
	return n.skiplist.remove(n, n.Key)
}

func (n *Node) _next() *Node {
	return n.nexts[0]
}

func (n *Node) Next() *Node {
	if m := n.nexts[0]; m != n.skiplist.terminus {
		return m
	} else {
		return nil
	}
}

func (n *Node) Prev() *Node {
	if m := n.prev; m != n.skiplist.terminus {
		return m
	} else {
		return nil
	}
}

func (n *Node) Reposition(k Comparable) {
	n.skiplist.reposition(n, k)
}

func (n *Node) nullify() {
	// this is called when n is removed from the skiplist. It's really
	// just to ensure that if someone has a reference to n lying
	// around, they can't use it.
	n.prev = nil
	n.nexts = nil
	n.skiplist = nil
}

func (s *SkipList) String() string {
	strs := make([]string, 1, s.length+1)
	strs[0] = fmt.Sprint(s.terminus)
	for cur := s.terminus._next(); cur != s.terminus; cur = cur._next() {
		strs = append(strs, fmt.Sprint(cur))
	}
	return fmt.Sprintf("Skiplist of length %v (counted: %v), levelProbabilities %v, and nodes:\n\t[%v]",
		s.length, len(strs)-1, s.levelProbabilities, strings.Join(strs, ",\n\t "))
}

func (n *Node) String() string {
	strs := make([]string, len(n.nexts))
	for idx := 0; idx < len(strs); idx++ {
		strs[idx] = fmt.Sprint(n.nexts[idx].Key)
	}
	return fmt.Sprintf("%v -> %v (nexts: [%v])", n.Key, n.Value, strings.Join(strs, ", "))
}
