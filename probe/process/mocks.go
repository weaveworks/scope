package process

// note: we must keep this in the "proc" package so it can be used from other packages..

// MockedReader is a mocked "/proc" reader
type MockedReader struct {
	Procs []Process
	Conns []Connection
}

// Read (mocked version)
func (mw *MockedReader) Read() error {
	return nil
}

// Processes walks through the processes provided in the mocked "/proc" reader
func (mw MockedReader) Processes(f func(Process)) error {
	for _, p := range mw.Procs {
		f(p)
	}
	return nil
}

// Connections walks through the connections provided in the mocked "/proc" reader
func (mw *MockedReader) Connections(f func(Connection)) error {
	for _, c := range mw.Conns {
		f(c)
	}
	return nil
}

// Close (mocked version)
func (mw *MockedReader) Close() error {
	return nil
}
