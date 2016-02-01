package skiplist

import (
	"math/rand"
	"os"
	"strconv"
	"testing"
)

type intKey int

func (sk intKey) LessThan(b Comparable) bool {
	return sk < b.(intKey)
}

func (sk intKey) Equal(b Comparable) bool {
	return sk == b.(intKey)
}

const (
	keyValsLen = 1000000
)

var (
	rng     = rand.New(rand.NewSource(0))
	keyVals = make(map[intKey]string)
	indices = make([]intKey, keyValsLen)
)

func TestAddFromEmptyRoot(t *testing.T) {
	for idx := 0; idx < 10; idx++ {
		s := New(rng)
		for idy := 0; idy < idx; idy++ {
			k := indices[idy]
			v := keyVals[k]
			n := s.Insert(k, v)
			if n.Key != k || n.Value != v {
				t.Fatal("Key or Value of inserted pair changed!", k, v, n)
			}
			if m := s.Get(k); n != m {
				t.Fatal("Node changed for just inserted value:", n, m)
			}
			if s.Len() != uint(idy+1) {
				t.Fatal("Incorrect length")
			}
		}
	}
}

func TestAddFromEmptyRel(t *testing.T) {
	for idx := 0; idx < 10; idx++ {
		s := New(rng)
		var n *Node
		for idy := 0; idy < idx; idy++ {
			k := indices[idy]
			v := keyVals[k]
			if n == nil {
				n = s.Insert(k, v)
			} else {
				n = n.Insert(k, v)
			}
			if n.Key != k || n.Value != v {
				t.Fatal("Key or Value of inserted pair changed!", k, v, n)
			}
			if m := n.Get(k); n != m {
				t.Fatal("Node changed for just inserted value:", n, m)
			}
			if p := n.Prev(); p != nil {
				if m := p.Get(k); n != m {
					t.Fatal("Node changed for just inserted value:", n, m)
				}
			}
			if s.Len() != uint(idy+1) {
				t.Fatal("Incorrect length")
			}
		}
	}
}

func TestDupInsert(t *testing.T) {
	s := New(rng)
	for idx := 0; idx < 10; idx++ {
		for idy := 0; idy < idx; idy++ {
			k := indices[idy]
			v := keyVals[k]
			if n := s.Insert(k, v); n.Key != k || n.Value != v {
				t.Fatal("Key or Value of inserted pair changed!", k, v, n)
			}
		}
	}
}

func TestGetMissingEmpty(t *testing.T) {
	s := New(rng)
	for idx := 0; idx < 10; idx++ {
		k := indices[idx]
		if n := s.Get(k); n != nil {
			t.Fatal("Expected not to find elem")
		}
	}
}

func TestGetMissingNonEmpty(t *testing.T) {
	s := New(rng)
	for idx := 0; idx < 20; idx++ {
		k := indices[idx]
		v := keyVals[k]
		if idx%2 == 0 {
			s.Insert(k, v)
		} else {
			if n := s.Get(k); n != nil {
				t.Fatal("Expected not to find elem")
			}
		}
	}
}

func TestRemoveRoot(t *testing.T) {
	s := New(rng)
	for idx := 0; idx < 20; idx++ {
		k := indices[idx]
		v := keyVals[k]
		s.Insert(k, v)
	}
	for idx := 0; idx < 20; idx++ {
		k := indices[idx]
		v := keyVals[k]
		if u := s.Remove(k); u != v {
			t.Fatal("Wrong value returned:", u, v)
		}
		if int(s.Len()) != 19-idx {
			t.Fatal("Wrong length")
		}
	}
}

func TestRemoveSelf(t *testing.T) {
	s := New(rng)
	ns := []*Node{}
	for idx := 0; idx < 20; idx++ {
		k := indices[idx]
		v := keyVals[k]
		ns = append(ns, s.Insert(k, v))
	}
	for idx, n := range ns {
		k := indices[idx]
		v := keyVals[k]
		if u := n.Remove(); u != v {
			t.Fatal("Wrong value returned:", u, v)
		}
		if int(s.Len()) != len(ns)-idx-1 {
			t.Fatal("Wrong length")
		}
	}
}

func TestRemoveMissing(t *testing.T) {
	s := New(rng)
	for idx := 0; idx < 20; idx++ {
		k := indices[idx]
		v := keyVals[k]
		s.Insert(k, v)
	}
	for idx := 0; idx < 20; idx++ {
		for idy := 0; idy < idx; idy++ {
			k := indices[idy]
			v := keyVals[k]
			u := s.Remove(k)
			if idy+1 == idx {
				if u != v {
					t.Fatal("Wrong value returned:", u, v)
				}
			} else {
				if u != nil {
					t.Fatal("Wrong value returned - expected nil:", u)
				}
			}
		}
	}
	if s.Len() != 1 {
		t.Fatal("Wrong length")
	}
}

func TestFirstLast(t *testing.T) {
	s := New(rng)
	if s.First() != nil || s.Last() != nil {
		t.Fatal("Expected nil for First and Last on empty list")
	}
	var min, max intKey
	for idx := 0; idx < 200; idx++ {
		k := indices[idx]
		v := keyVals[k]
		s.Insert(k, v)
		if idx == 0 {
			min, max = k, k
		} else {
			if k < min {
				min = k
			}
			if k > max {
				max = k
			}
		}
	}
	if f := s.First(); f.Key != min {
		t.Fatal("Did not get minimum key back for first", min, f)
	}
	if l := s.Last(); l.Key != max {
		t.Fatal("Did not get maximum key back for last", max, l)
	}
}

func TestMerge(t *testing.T) {
	s0 := New(rng)
	s1 := New(rng)
	lim := 1000
	for idx := 0; idx < lim; idx++ {
		if idx%2 == 0 {
			s0.Insert(intKey(idx), idx)
		} else {
			s1.Insert(intKey(idx), idx)
		}
	}
	s0.Merge(s1)
	if int(s0.Len()) != lim {
		t.Fatal("Wrong len after merge", s0.Len())
	}
	cur := s0.First()
	for idx := 0; idx < lim; idx++ {
		if cur.Value.(int) != idx {
			t.Fatal("Wrong value: ", cur.Value)
		}
		if cur != s0.Get(cur.Key) {
			t.Fatal("Internal failure: ", cur)
		}
		cur = cur.Next()
	}
}

func TestReposition(t *testing.T) {
	s := New(rng)
	lim := 200
	for idx := 0; idx < lim; idx++ {
		s.Insert(intKey(idx*5), idx)
	}
	for idx := lim - 1; idx >= 0; idx-- {
		n := s.Get(intKey(idx * 5))
		if n == nil {
			t.Fatal("Unable to find node")
		}
		n.Reposition(intKey((idx * 5) - 11))
	}
	n := s.First()
	for idx := 0; idx < lim; idx++ {
		if n.Value != idx {
			t.Fatal("Wrong value", idx, n)
		}
		n = n.Next()
	}
	if n != nil {
		t.Fatal("Too many values!")
	}
}

func BenchmarkAdd08192(b *testing.B) { benchmarkAdd(8192, b) }
func BenchmarkAdd16384(b *testing.B) { benchmarkAdd(16384, b) }
func BenchmarkAdd32768(b *testing.B) { benchmarkAdd(32768, b) }
func BenchmarkAdd65536(b *testing.B) { benchmarkAdd(65536, b) }

func benchmarkAdd(initial int, b *testing.B) {
	s := populate(New(rng), 0, initial)
	b.ResetTimer()
	populateFast(s, initial, b.N)
}

func BenchmarkGet08192(b *testing.B) { benchmarkGet(8192, b) }
func BenchmarkGet16384(b *testing.B) { benchmarkGet(16384, b) }
func BenchmarkGet32768(b *testing.B) { benchmarkGet(32768, b) }
func BenchmarkGet65536(b *testing.B) { benchmarkGet(65536, b) }

func benchmarkGet(initial int, b *testing.B) {
	s := populate(New(rng), 0, initial)
	b.ResetTimer()
	for idx := 0; idx < b.N; idx++ {
		s.Get(indices[idx%initial])
	}
}

func populate(s *SkipList, offset, lim int) *SkipList {
	for idx := 0; idx < lim; idx++ {
		k := indices[(offset+idx)%len(indices)]
		v := keyVals[k]
		s.Insert(k, v)
	}
	return s
}

func populateFast(s *SkipList, offset, lim int) *SkipList {
	for idx := 0; idx < lim; idx++ {
		n := idx + offset
		s.Insert(intKey(n), n)
	}
	return s
}

func TestMain(m *testing.M) {
	for idx := 0; idx < keyValsLen; idx++ {
		keyVals[intKey(idx)] = strconv.FormatInt(int64(idx), 3)
	}
	for idx, k := range rand.Perm(keyValsLen) {
		indices[idx] = intKey(k)
	}
	os.Exit(m.Run())
}
