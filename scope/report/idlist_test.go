package report

import (
	"reflect"
	"testing"
)

func TestidList(t *testing.T) {
	var have IDList
	have = have.Add("aap")
	have = have.Add("noot")
	have = have.Add("mies")
	have = have.Add("aap")
	have = have.Add("aap")
	have = have.Add("wim")
	have = have.Add("vuur")

	if want := IDList([]string{"aap", "mies", "noot", "vuur", "wim"}); !reflect.DeepEqual(want, have) {
		t.Errorf("want %+v, have %+v", want, have)
	}
}
