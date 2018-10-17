package cri_test

import (
	"testing"

	"github.com/bmizerany/assert"
	"github.com/weaveworks/scope/probe/cri"
)

func TestParseHttpEndpointUrl(t *testing.T) {
	_, err := cri.NewCRIClient("http://xyz.com")

	assert.Equal(t, "protocol \"http\" not supported", err.Error())
}

func TestParseTcpEndpointUrl(t *testing.T) {
	client, err := cri.NewCRIClient("127.0.0.1")

	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, client)
}
