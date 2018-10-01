package cri_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/cri"
)

func TestParseHttpEndpointUrl(t *testing.T) {
	_, err := cri.NewCRIClient("http://xyz.com")

	if err == nil {
		t.Fatal("Should not create client with Http protocol")
	}
}

func TestParseTcpEndpointUrl(t *testing.T) {
	client, err := cri.NewCRIClient("127.0.0.1")

	if err != nil {
		t.Fatal("Should have created service client")
	}

	if client == nil {
		t.Fatal("Should have created service client")
	}
}
