package process_test

import (
	"testing"

	"github.com/weaveworks/scope/probe/process"
)

func TestProcReaderBasic(t *testing.T) {
	procFunc := func(process.Process) {}
	nullProcDir := mockedDir{
		Dir:              "",
		OpenFunc:         func(string) (process.File, error) { return &process.OSFile{}, nil },
		ReadDirNamesFunc: func(string) ([]string, error) { return []string{}, nil },
	}

	reader := process.NewReader(nullProcDir, false)
	if err := reader.Read(); err != nil {
		t.Fatal(err)
	}
	if err := reader.Processes(procFunc); err != nil {
		t.Fatal(err)
	}
}
