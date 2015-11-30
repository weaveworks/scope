package procspy

import (
	"bytes"
	"testing"
)

func BenchmarkParseConnectionsBaseline(b *testing.B) {
	readFile = func(string, *bytes.Buffer) error { return nil }
	benchmarkConnections(b)
	// 333 ns/op, 0 allocs/op
}

func BenchmarkParseConnectionsFixture(b *testing.B) {
	readFile = func(_ string, buf *bytes.Buffer) error { _, err := buf.Write(fixture); return err }
	benchmarkConnections(b)
	// 15553 ns/op, 12 allocs/op
}

func benchmarkConnections(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cbConnections(false)
	}
}

var fixture = []byte(`  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:A6C0 00000000:0000 01 00000000:00000000 00:00000000 00000000   105        0 5107 1 ffff8800a6aaf040 100 0 0 10 0
   1: 00000000:006F 00000000:0000 01 00000000:00000000 00:00000000 00000000     0        0 5084 1 ffff8800a6aaf740 100 0 0 10 0
   2: 0100007F:0019 00000000:0000 01 00000000:00000000 00:00000000 00000000     0        0 10550 1 ffff8800a729b780 100 0 0 10 0
   3: A12CF62E:E4D7 57FC1EC0:01BB 01 00000000:00000000 02:000006FA 00000000  1000        0 639474 2 ffff88007e75a740 48 4 26 10 -1
`)
