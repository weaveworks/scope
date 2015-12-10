package procspy

// SetFixtures declares constant Connection and ConnectionProcs which will
// always be returned by the package-level Connections and Processes
// functions. It's designed to be used in tests.

type fixedConnIter []Connection

func (f *fixedConnIter) Next() *Connection {
	if len(*f) == 0 {
		return nil
	}

	car := (*f)[0]
	*f = (*f)[1:]

	return &car
}

// SetFixtures is used in test scenarios to have known output.
func SetFixtures(c []Connection) {
	cbConnections = func(bool) (ConnIter, error) {
		f := fixedConnIter(c)
		return &f, nil
	}
}
