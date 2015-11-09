package ps

import "testing"

func TestListImmutable(t *testing.T) {
	// build some lists
	one := NewList().Cons("first")
	two := one.Cons("second")
	zwei := one.Cons("zweite")

	// check each list's length
	if size := one.Size(); size != 1 {
		t.Errorf("one doesn't have 1 item, it has %d", size)
	}
	if size := two.Size(); size != 2 {
		t.Errorf("two doesn't have 2 items, it has %d", size)
	}
	if size := zwei.Size(); size != 2 {
		t.Errorf("zwei doesn't have 2 item, it has %d", size)
	}

	// check each list's contents
	if one.Head() != "first" {
		t.Errorf("one has the wrong head")
	}
	if two.Head() != "second" {
		t.Errorf("two has the wrong head")
	}
	if two.Tail().Head() != "first" {
		t.Errorf("two has the wrong ending")
	}
	if zwei.Head() != "zweite" {
		t.Errorf("zwei has the wrong head")
	}
	if zwei.Tail().Head() != "first" {
		t.Errorf("zwei has the wrong ending")
	}
}

// benchmark making a really long list
func BenchmarkListCons(b *testing.B) {
	l := NewList()
	for i := 0; i < b.N; i++ {
		l = l.Cons(i)
	}
}
