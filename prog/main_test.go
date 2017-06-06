package main

import (
	"flag"
	"testing"

	"github.com/Sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func TestMakeContainerFiltersFromFlags(t *testing.T) {
	containerLabelFlags := containerLabelFiltersFlag{exclude: false}
	containerLabelFlags.Set(`title1:label=1`)
	containerLabelFlags.Set(`ti\:tle2:lab\:el=2`)
	containerLabelFlags.Set(`ti tile3:label=3`)

	err := containerLabelFlags.Set(`just a string`)
	assert.NotNil(t, err, "Invalid container label flag not detected")

	apiTopologyOptions := containerLabelFlags.apiTopologyOptions
	assert.Equal(t, 3, len(apiTopologyOptions))
	assert.Equal(t, `title1`, apiTopologyOptions[0].Value)
	assert.Equal(t, `label=1`, apiTopologyOptions[0].Label)
	assert.Equal(t, `ti:tle2`, apiTopologyOptions[1].Value)
	assert.Equal(t, `lab:el=2`, apiTopologyOptions[1].Label)
	assert.Equal(t, `ti tle3`, apiTopologyOptions[2].Value)
	assert.Equal(t, `label=3`, apiTopologyOptions[2].Label)
}

func TestLogCensoredArgs(t *testing.T) {
	setupFlags(&flags{})
	args := []string{
		"-probe.token=secret",
		"-service-token=secret",
		"-probe.kubernetes.password=secret",
		"-probe.kubernetes.token=secret",
		"http://secret:secret@frontend.dev.weave.works:80",
		"https://secret:secret@cloud.weave.works:443",
	}
	flag.CommandLine.Parse(args)

	hook := test.NewGlobal()
	logCensoredArgs()
	assert.NotContains(t, hook.LastEntry().Message, "secret")
	assert.Contains(t, hook.LastEntry().Message, "cloud.weave.works:443")
}
