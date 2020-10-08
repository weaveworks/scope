package cri_test

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/weaveworks/scope/probe/cri"
)

var nonUnixSocketsTest = []struct {
	endpoint     string
	errorMessage string
}{
	{"http://xyz.com", "protocol \"http\" not supported"},
	{"tcp://var/unix.sock", "endpoint was not unix socket tcp"},
	{"http://[fe80::%31]/", "parse \"http://[fe80::%31]/\": invalid URL escape \"%31\""},
}

func TestParseNonUnixEndpointUrl(t *testing.T) {
	for _, tt := range nonUnixSocketsTest {
		_, err := cri.NewCRIClient(tt.endpoint)

		assert.Equal(t, tt.errorMessage, err.Error())
	}
}

var unixSocketsTest = []string{
	"127.0.0.1", // tests the fallback endpoint
	"unix://127.0.0.1",
	"unix///var/run/dockershim.sock",
	"var/run/dockershim.sock",
}

func TestParseUnixEndpointUrl(t *testing.T) {
	for _, tt := range unixSocketsTest {
		client, err := cri.NewCRIClient(tt)

		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, client)
	}

}
