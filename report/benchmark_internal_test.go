package report

import (
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var (
	benchReportPath = flag.String("bench-report-path", "", "report file, or dir with files, to use for benchmarking (relative to this package)")
)

func BenchmarkReportUnmarshal(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		b.StartTimer()
		if err := readReportFiles(*benchReportPath); err != nil {
			b.Fatal(err)
		}
	}
}

func readReportFiles(path string) error {
	return filepath.Walk(path,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if _, err := MakeFromFile(p); err != nil {
				return err
			}
			return nil
		})
}
