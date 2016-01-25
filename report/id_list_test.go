package report_test

import (
	"testing"

	"github.com/weaveworks/scope/report"
	"github.com/weaveworks/scope/test/reflect"
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
