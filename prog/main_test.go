package main_test

import (
	"fmt"
	"github.com/weaveworks/scope/app"
	"testing"
)

func TestMakeContainerFiltersFromFlags(t *testing.T) {
	containerLabelFlags := containerLabelFiltersFlag{exclude: false}
	containerLabelFlags.Set(`title1:label=1`)
	containerLabelFlags.Set(`ti\:tle2:lab\:el=2`)
	containerLabelFlags.Set(`ti tile3:label=3`)
	err := containerLabelFlags.Set(`just a string`)
	if err == nil {
		t.Fatalf("Invalid container label flag not detected")
	}
	apiTopologyOptions := containerLabelFlags.apiTopologyOptions
	equals(t, 3, len(apiTopologyOptions))
	equals(t, `title1`, apiTopologyOptions[0].Value)
	equals(t, `label=1`, apiTopologyOptions[0].Label)
	equals(t, `ti:tle2`, apiTopologyOptions[1].Value)
	equals(t, `lab:el=2`, apiTopologyOptions[1].Label)
	equals(t, `ti tle3`, apiTopologyOptions[2].Value)
	equals(t, `label=3`, apiTopologyOptions[2].Label)
}
