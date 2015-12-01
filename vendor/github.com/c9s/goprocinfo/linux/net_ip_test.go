package linux

import (
	"testing"
)

func TestNetIPv4DecoderRemote(t *testing.T) {

	ip, err := NetIPv4Decoder("00000000:0050")

	if err != nil {
		t.Fatal("net ipv4 decode fail", err)
	}

	if ip != "0.0.0.0:80" {
		t.Error("unexpected value")
	}

	t.Logf("%+v", ip)
}

func TestNetIPv4DecoderLocal(t *testing.T) {

	ip, err := NetIPv4Decoder("0100007F:1F90")

	if err != nil {
		t.Fatal("net ipv4 decode fail", err)
	}

	if ip != "127.0.0.1:8080" {
		t.Error("unexpected value")
	}

	t.Logf("%+v", ip)
}

func TestNetIPv6DecoderRemote(t *testing.T) {

	ip, err := NetIPv6Decoder("350E012A900F122E85EDEAADA64DAAD1:0016")

	if err != nil {
		t.Fatal("net ipv6 decode fail", err)
	}

	if ip != "2a01:e35:2e12:f90:adea:ed85:d1aa:4da6:22" {
		t.Error("unexpected value")
	}

	t.Logf("%+v", ip)
}

func TestNetIPv6DecoderLocal(t *testing.T) {

	ip, err := NetIPv6Decoder("00000000000000000000000001000000:2328")

	if err != nil {
		t.Fatal("net ipv6 decode fail", err)
	}

	if ip != "::1:9000" {
		t.Error("unexpected value")
	}

	t.Logf("%+v", ip)
}
