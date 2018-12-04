package speed

// Instances defines a valid collection of instance name and values
type Instances map[string]interface{}

// Keys collects and returns all the keys in all instance values
func (i Instances) Keys() []string {
	s := make([]string, 0, len(i))
	for k := range i {
		s = append(s, k)
	}
	return s
}

// pcpInstance wraps a PCP compatible Instance
type pcpInstance struct {
	name   string
	id     uint32
	offset int
}

// newpcpInstance generates a new Instance type based on the passed parameters
// the id is passed explicitly as it is assumed that this will be constructed
// after initializing the InstanceDomain
// this is not a part of the public API as this is not supposed to be used directly,
// but instead added using the AddInstance method of InstanceDomain
func newpcpInstance(name string) *pcpInstance {
	return &pcpInstance{
		name, hash(name, 0), 0,
	}
}
