package report_test

import (
	"reflect"
	"testing"

	"github.com/weaveworks/scope/report"
)

func TestIDList(t *testing.T) {
	have := report.MakeIDList("alpha", "mu", "zeta")
	have = have.Add("alpha")
	have = have.Add("nu")
	have = have.Add("mu")
	have = have.Add("alpha")
	have = have.Add("alpha")
	have = have.Add("epsilon")
	have = have.Add("delta")
	if want := report.IDList([]string{"alpha", "delta", "epsilon", "mu", "nu", "zeta"}); !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}

func TestIDListCopy(t *testing.T) {
	one := report.MakeIDList("a", "b", "c")
	two := one.Copy()
	one.Add("d")
	if want, have := report.MakeIDList("a", "b", "c"), two; !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}
