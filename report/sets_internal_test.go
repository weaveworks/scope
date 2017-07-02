package report

import (
	"testing"

	"github.com/weaveworks/scope/test/reflect"
)

func TestSets(t *testing.T) {
	sets := MakeSets().Add("foo", MakeStringSet("bar"))
	if v, _ := sets.Lookup("foo"); !reflect.DeepEqual(v, MakeStringSet("bar")) {
		t.Fatal(v)
	}

	sets = sets.Merge(MakeSets().Add("foo", MakeStringSet("baz")))
	if v, _ := sets.Lookup("foo"); !reflect.DeepEqual(v, MakeStringSet("bar", "baz")) {
		t.Fatal(v)
	}
}
