// Package mmvdump implements a go port of the C mmvdump utility included in PCP Core
//
// https://github.com/performancecopilot/pcp/blob/master/src/pmdas/mmv/mmvdump.c
//
// It has been written for maximum portability with the C equivalent, without having to use cgo or any other ninja stuff
//
// the main difference is that the reader is separate from the cli with the reading primarily implemented in mmvdump.go while the cli is implemented in cmd/mmvdump
//
// the cli application is completely go gettable and outputs the same things, in mostly the same way as the C cli app, to try it out,
//
// ```
// go get github.com/performancecopilot/speed/mmvdump/cmd/mmvdump
// ```
package mmvdump

import (
	"math"
	"sync"
	"unsafe"

	"github.com/pkg/errors"
)

func readHeader(data []byte) (*Header, error) {
	if uint64(len(data)) < HeaderLength {
		return nil, errors.New("file too small to contain a valid Header")
	}

	header := (*Header)(unsafe.Pointer(&data[0]))

	if m := header.Magic[:3]; string(m) != "MMV" {
		return nil, errors.Errorf("Bad Magic: %v", string(m))
	}

	if header.G1 != header.G2 {
		return nil, errors.Errorf("Mismatched version numbers, %v and %v", header.G1, header.G2)
	}

	return header, nil
}

func readToc(data []byte, offset uint64) (*Toc, error) {
	if uint64(len(data)) < offset+TocLength {
		return nil, errors.New("Incomplete/Partially Written TOC")
	}

	return (*Toc)(unsafe.Pointer(&data[offset])), nil
}

type itemReaderFunc func([]byte, uint64, int32) (interface{}, error)

func readInstance(data []byte, offset uint64, version int32) (interface{}, error) {
	var InstanceLength = Instance1Length
	if version == 2 {
		InstanceLength = Instance2Length
	}

	if uint64(len(data)) < offset+InstanceLength {
		return nil, errors.New("Incomplete/Partially Written Instance")
	}

	if version == 1 {
		return (*Instance1)(unsafe.Pointer(&data[offset])), nil
	}

	return (*Instance2)(unsafe.Pointer(&data[offset])), nil
}

func readInstanceDomain(data []byte, offset uint64, version int32) (interface{}, error) {
	if uint64(len(data)) < offset+InstanceDomainLength {
		return nil, errors.New("Incomplete/Partially Written InstanceDomain")
	}

	return (*InstanceDomain)(unsafe.Pointer(&data[offset])), nil
}

func readMetric(data []byte, offset uint64, version int32) (interface{}, error) {
	var MetricLength = Metric1Length
	if version == 2 {
		MetricLength = Metric2Length
	}

	if uint64(len(data)) < offset+MetricLength {
		return nil, errors.New("Incomplete/Partially Written Metric")
	}

	if version == 1 {
		return (*Metric1)(unsafe.Pointer(&data[offset])), nil
	}

	return (*Metric2)(unsafe.Pointer(&data[offset])), nil
}

func readValue(data []byte, offset uint64, version int32) (interface{}, error) {
	if uint64(len(data)) < offset+ValueLength {
		return nil, errors.New("Incomplete/Partially Written Value")
	}

	return (*Value)(unsafe.Pointer(&data[offset])), nil
}

func readString(data []byte, offset uint64, version int32) (interface{}, error) {
	if uint64(len(data)) < offset+StringLength {
		return nil, errors.New("Incomplete/Partially Written String")
	}

	return (*String)(unsafe.Pointer(&data[offset])), nil
}

func readTocs(data []byte, count int32) ([]*Toc, error) {
	tocs := make([]*Toc, count)

	for i := int32(0); i < count; i++ {
		t, err := readToc(data, HeaderLength+uint64(i)*TocLength)
		if err != nil {
			return nil, err
		}
		tocs[i] = t
	}

	return tocs, nil
}

func readItems(data []byte, offset uint64, count int32, itemlength uint64, readItem itemReaderFunc, version int32) (map[uint64]interface{}, error) {
	var wg sync.WaitGroup
	wg.Add(int(count))

	items := make(map[uint64]interface{})

	var (
		err error
		m   sync.Mutex
	)

	for i := int32(0); i < count; i, offset = i+1, offset+itemlength {
		go func(offset uint64) {
			if err == nil {
				item, ierr := readItem(data, offset, version)
				if ierr == nil {
					m.Lock()
					items[offset] = item
					m.Unlock()
				} else {
					err = ierr
				}
			}
			wg.Done()
		}(offset)
	}

	wg.Wait()

	if err != nil {
		return nil, err
	}

	return items, nil
}

func readInstances(data []byte, offset uint64, count int32, version int32) (map[uint64]Instance, error) {
	InstanceLength := Instance1Length
	if version == 2 {
		InstanceLength = Instance2Length
	}

	i, err := readItems(data, offset, count, InstanceLength, readInstance, version)
	if err != nil {
		return nil, err
	}

	instances := make(map[uint64]Instance)
	for off, val := range i {
		instances[off] = val.(Instance)
	}

	return instances, nil
}

func readInstanceDomains(data []byte, offset uint64, count int32, version int32) (map[uint64]*InstanceDomain, error) {
	i, err := readItems(data, offset, count, InstanceDomainLength, readInstanceDomain, version)
	if err != nil {
		return nil, err
	}

	indoms := make(map[uint64]*InstanceDomain)
	for off, val := range i {
		indoms[off] = val.(*InstanceDomain)
	}

	return indoms, nil
}

func readMetrics(data []byte, offset uint64, count int32, version int32) (map[uint64]Metric, error) {
	var MetricLength = Metric1Length
	if version == 2 {
		MetricLength = Metric2Length
	}

	m, err := readItems(data, offset, count, MetricLength, readMetric, version)
	if err != nil {
		return nil, err
	}

	metrics := make(map[uint64]Metric)
	for off, val := range m {
		metrics[off] = val.(Metric)
	}

	return metrics, nil
}

func readValues(data []byte, offset uint64, count int32, version int32) (map[uint64]*Value, error) {
	v, err := readItems(data, offset, count, ValueLength, readValue, version)
	if err != nil {
		return nil, err
	}

	values := make(map[uint64]*Value)
	for off, val := range v {
		values[off] = val.(*Value)
	}

	return values, nil
}

func readStrings(data []byte, offset uint64, count int32, version int32) (map[uint64]*String, error) {
	s, err := readItems(data, offset, count, StringLength, readString, version)
	if err != nil {
		return nil, err
	}

	strings := make(map[uint64]*String)
	for off, val := range s {
		strings[off] = val.(*String)
	}

	return strings, nil
}

func readComponents(data []byte, tocs []*Toc, version int32) (
	metrics map[uint64]Metric,
	values map[uint64]*Value,
	instances map[uint64]Instance,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
	ierr, inerr, merr, verr, serr error,
) {
	var wg sync.WaitGroup
	wg.Add(len(tocs))

	for _, toc := range tocs {
		switch toc.Type {
		case TocInstances:
			go func(offset uint64, count int32) {
				instances, ierr = readInstances(data, offset, count, version)
				wg.Done()
			}(toc.Offset, toc.Count)
		case TocIndoms:
			go func(offset uint64, count int32) {
				indoms, inerr = readInstanceDomains(data, offset, count, version)
				wg.Done()
			}(toc.Offset, toc.Count)
		case TocMetrics:
			go func(offset uint64, count int32) {
				metrics, merr = readMetrics(data, offset, count, version)
				wg.Done()
			}(toc.Offset, toc.Count)
		case TocValues:
			go func(offset uint64, count int32) {
				values, verr = readValues(data, offset, count, version)
				wg.Done()
			}(toc.Offset, toc.Count)
		case TocStrings:
			go func(offset uint64, count int32) {
				strings, serr = readStrings(data, offset, count, version)
				wg.Done()
			}(toc.Offset, toc.Count)
		}
	}

	wg.Wait()

	return
}

// Dump creates a data dump from the passed data
func Dump(data []byte) (
	h *Header,
	tocs []*Toc,
	metrics map[uint64]Metric,
	values map[uint64]*Value,
	instances map[uint64]Instance,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
	err error,
) {
	h, err = readHeader(data)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	tocs, err = readTocs(data, h.Toc)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, nil, err
	}

	var ierr, inerr, merr, verr, serr error

	metrics, values, instances, indoms, strings, ierr, inerr, merr, verr, serr = readComponents(data, tocs, h.Version)

	switch {
	case ierr != nil:
		return nil, nil, nil, nil, nil, nil, nil, ierr
	case inerr != nil:
		return nil, nil, nil, nil, nil, nil, nil, inerr
	case merr != nil:
		return nil, nil, nil, nil, nil, nil, nil, merr
	case verr != nil:
		return nil, nil, nil, nil, nil, nil, nil, verr
	case serr != nil:
		return nil, nil, nil, nil, nil, nil, nil, serr
	}

	return
}

// FixedVal will infer a fixed size value from the passed data
func FixedVal(data uint64, t Type) (interface{}, error) {
	switch t {
	case Int32Type:
		return int32(data), nil
	case Uint32Type:
		return uint32(data), nil
	case Int64Type:
		return int64(data), nil
	case Uint64Type:
		return data, nil
	case FloatType:
		return math.Float32frombits(uint32(data)), nil
	case DoubleType:
		return math.Float64frombits(data), nil
	}

	return nil, errors.New("invalid type")
}
