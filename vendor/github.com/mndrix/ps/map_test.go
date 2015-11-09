package ps

import . "strconv"

import "testing"
import "sort"

func TestMapNil(t *testing.T) {
	m := NewMap()
	keys := m.Keys()
	if len(keys) != 0 {
		t.Errorf("Empty map has keys")
	}
}

func TestMapImmutable(t *testing.T) {
	// build a couple small maps
	world := NewMap().Set("hello", "world")
	kids := world.Set("hello", "kids")

	// both maps should still retain their data
	if v, _ := world.Lookup("hello"); v != "world" {
		t.Errorf("Set() modified the receiving map")
	}
	if size := world.Size(); size != 1 {
		t.Errorf("world size is not 1 : %d", size)
	}
	if v, _ := kids.Lookup("hello"); v != "kids" {
		t.Errorf("Set() did not modify the resulting map")
	}
	if size := kids.Size(); size != 1 {
		t.Errorf("kids size is not 1 : %d", size)
	}

	// both maps have the right keys
	if keys := world.Keys(); len(keys) != 1 || keys[0] != "hello" {
		t.Errorf("world has the wrong keys: %#v", keys)
	}
	if keys := kids.Keys(); len(keys) != 1 || keys[0] != "hello" {
		t.Errorf("kids has the wrong keys: %#v", keys)
	}

	// test deletion
	empty := kids.Delete("hello")
	if size := empty.Size(); size != 0 {
		t.Errorf("empty size is not 1 : %d", size)
	}
	if keys := empty.Keys(); len(keys) != 0 {
		t.Errorf("empty has the wrong keys: %#v", keys)
	}
}

func TestMapMultipleKeys(t *testing.T) {
	// map with multiple keys each with pointer values
	one := 1
	two := 2
	three := 3
	m := NewMap().Set("one", &one).Set("two", &two).Set("three", &three)

	// do we have the right number of keys?
	keys := m.Keys()
	if len(keys) != 3 {
		t.Logf("wrong size keys: %d", len(keys))
		t.FailNow()
	}

	// do we have the right keys?
	sort.Strings(keys)
	if keys[0] != "one" {
		t.Errorf("unexpected key: %s", keys[0])
	}
	if keys[1] != "three" {
		t.Errorf("unexpected key: %s", keys[1])
	}
	if keys[2] != "two" {
		t.Errorf("unexpected key: %s", keys[2])
	}

	// do we have the right values?
	vp, ok := m.Lookup("one")
	if !ok {
		t.Logf("missing value for one")
		t.FailNow()
	}
	if v := vp.(*int); *v != 1 {
		t.Errorf("wrong value: %d\n", *v)
	}
	vp, ok = m.Lookup("two")
	if !ok {
		t.Logf("missing value for two")
		t.FailNow()
	}
	if v := vp.(*int); *v != 2 {
		t.Errorf("wrong value: %d\n", *v)
	}
	vp, ok = m.Lookup("three")
	if !ok {
		t.Logf("missing value for three")
		t.FailNow()
	}
	if v := vp.(*int); *v != 3 {
		t.Errorf("wrong value: %d\n", *v)
	}
}

func TestMapManyKeys(t *testing.T) {
	// build a map with many keys and values
	count := 100
	m := NewMap()
	for i := 0; i < count; i++ {
		m = m.Set(Itoa(i), i)
	}

	if m.Size() != 100 {
		t.Errorf("Wrong number of keys: %d", m.Size())
	}

	m = m.Delete("42").Delete("7").Delete("19").Delete("99")
	if m.Size() != 96 {
		t.Errorf("Wrong number of keys: %d", m.Size())
	}

	for i := 43; i < 99; i++ {
		v, ok := m.Lookup(Itoa(i))
		if !ok || v != i {
			t.Errorf("Wrong value for key %d", i)
		}
	}
}

func TestMapHashKey(t *testing.T) {
	hash := hashKey("this is a key")
	if hash != 10424450902216330915 {
		t.Errorf("This isn't FNV-1a hashing: %d", hash)
	}
}

func BenchmarkMapSet(b *testing.B) {
	m := NewMap()
	for i := 0; i < b.N; i++ {
		m = m.Set("foo", i)
	}
}

func BenchmarkMapDelete(b *testing.B) {
	m := NewMap().Set("key", "value")
	for i := 0; i < b.N; i++ {
		m.Delete("key")
	}
}

func BenchmarkHashKey(b *testing.B) {
	key := "this is a key"
	for i := 0; i < b.N; i++ {
		_ = hashKey(key)
	}
}
