package report

import (
	"testing"

	"$GITHUB_URI/test/reflect"
)

func TestSets(t *testing.T) {
	sets := EmptySets.Add("foo", MakeStringSet("bar"))
	if v, _ := sets.Lookup("foo"); !reflect.DeepEqual(v, MakeStringSet("bar")) {
		t.Fatal(v)
	}

	sets = sets.Merge(EmptySets.Add("foo", MakeStringSet("baz")))
	if v, _ := sets.Lookup("foo"); !reflect.DeepEqual(v, MakeStringSet("bar", "baz")) {
		t.Fatal(v)
	}
}
