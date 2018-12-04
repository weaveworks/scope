package mmvdump

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func instanceName(m Instance, header *Header, strings map[uint64]*String) string {
	if header.Version == 1 {
		return string(m.(*Instance1).External[:])
	}
	return string(strings[m.(*Instance2).External].Payload[:])
}

func writeInstance(
	w io.Writer,
	offset uint64,
	header *Header,
	instances map[uint64]Instance,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
) error {
	i := instances[offset]
	indom := indoms[i.Indom()]
	Name := instanceName(i, header, strings)
	_, err := fmt.Fprintf(w, "\t[%v/%v] instance = [%v/%v]\n", indom.Serial, offset, i.Internal(), Name)
	return err
}

func writeInstanceDomain(
	w io.Writer,
	offset uint64,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
) error {
	indom := indoms[offset]
	_, err := fmt.Fprintf(w, "\t[%v/%v] %d instances, starting at offset %d\n", indom.Serial, offset, indom.Count, indom.Offset)
	if err != nil {
		return err
	}

	if indom.Shorttext == 0 {
		if _, err := fmt.Fprintf(w, "\t\t(no shorttext)\n"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "\t\tshorttext=%v\n", string(strings[indom.Shorttext].Payload[:])); err != nil {
			return err
		}
	}

	if indom.Longtext == 0 {
		if _, err := fmt.Fprintf(w, "\t\t(no longtext)\n"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "\t\tlongtext=%v\n", string(strings[indom.Longtext].Payload[:])); err != nil {
			return err
		}
	}

	return nil
}

func metricName(m Metric, header *Header, strings map[uint64]*String) string {
	if header.Version == 1 {
		return string(m.(*Metric1).Name[:])
	}
	return string(strings[m.(*Metric2).Name].Payload[:])
}

func writeMetric(
	w io.Writer,
	offset uint64,
	header *Header,
	metrics map[uint64]Metric,
	strings map[uint64]*String,
) error {
	m := metrics[offset]
	Name := metricName(m, header, strings)

	if _, err := fmt.Fprintf(w, "\t[%v/%v] %v\n", m.Item(), offset, Name); err != nil {
		return err
	}

	_, err := fmt.Fprintf(w, "\t\ttype=%v (0x%x), sem=%v (0x%x), pad=0x%x\n", m.Typ(), int(m.Typ()), m.Sem(), int(m.Sem()), m.Padding())
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, "\t\tunits=%v\n", m.Unit()); err != nil {
		return err
	}

	if m.Indom() == NoIndom {
		if _, err := fmt.Fprintf(w, "\t\t(no indom)\n"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "\t\tindom=%d\n", m.Indom()); err != nil {
			return err
		}
	}

	if m.ShortText() == 0 {
		if _, err := fmt.Fprintf(w, "\t\t(no shorttext)\n"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "\t\tshorttext=%v\n", string(strings[m.ShortText()].Payload[:])); err != nil {
			return err
		}
	}

	if m.LongText() == 0 {
		if _, err := fmt.Fprintf(w, "\t\t(no longtext)\n"); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(w, "\t\tlongtext=%v\n", string(strings[m.LongText()].Payload[:])); err != nil {
			return err
		}
	}

	return nil
}

func writeValue(
	w io.Writer,
	offset uint64,
	header *Header,
	metrics map[uint64]Metric,
	values map[uint64]*Value,
	instances map[uint64]Instance,
	strings map[uint64]*String,
) error {
	v := values[offset]
	m := metrics[v.Metric]

	if _, err := fmt.Fprintf(w, "\t[%v/%v] %v", m.Item(), offset, metricName(m, header, strings)); err != nil {
		return err
	}

	var (
		a   interface{}
		err error
	)

	if m.Typ() != StringType {
		a, err = FixedVal(v.Val, m.Typ())
		if err != nil {
			return err
		}
	} else {
		v, ok := strings[uint64(v.Extra)]
		if !ok {
			return errors.Errorf("invalid string address")
		}
		a = string(v.Payload[:])
	}

	if m.Indom() != NoIndom && m.Indom() != 0 {
		i := instances[v.Instance]
		if _, err := fmt.Fprintf(w, "[%d or \"%s\"]", i.Internal(), instanceName(i, header, strings)); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(w, " = %v\n", a)
	return err
}

func writeString(w io.Writer, offset uint64, strings map[uint64]*String) error {
	_, err := fmt.Fprintf(w, "\t[%v] %v\n", offset, string(strings[offset].Payload[:]))
	return err
}

func writeComponents(
	w io.Writer,
	header *Header,
	tocs []*Toc,
	metrics map[uint64]Metric,
	values map[uint64]*Value,
	instances map[uint64]Instance,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
) error {
	var (
		toff                         = HeaderLength
		itemtype                     string
		itemsize                     uint64
		writeItem                    func(uint64) error
		InstanceLength, MetricLength uint64
	)

	if header.Version == 1 {
		InstanceLength = Instance1Length
		MetricLength = Metric1Length
	} else {
		InstanceLength = Instance2Length
		MetricLength = Metric2Length
	}

	for ti, toc := range tocs {
		switch toc.Type {
		case TocInstances:
			itemtype = "instances"
			itemsize = InstanceLength
			writeItem = func(off uint64) error { return writeInstance(w, off, header, instances, indoms, strings) }
		case TocIndoms:
			itemtype = "indoms"
			itemsize = InstanceDomainLength
			writeItem = func(off uint64) error { return writeInstanceDomain(w, off, indoms, strings) }
		case TocMetrics:
			itemtype = "metric"
			itemsize = MetricLength
			writeItem = func(off uint64) error { return writeMetric(w, off, header, metrics, strings) }
		case TocValues:
			itemtype = "values"
			itemsize = ValueLength
			writeItem = func(off uint64) error { return writeValue(w, off, header, metrics, values, instances, strings) }
		case TocStrings:
			itemtype = "strings"
			itemsize = StringLength
			writeItem = func(off uint64) error { return writeString(w, off, strings) }
		}

		if _, err := fmt.Fprintf(w, "TOC[%v], offset: %v, %v offset: %v (%v entries)\n", ti, toff, itemtype, toc.Offset, toc.Count); err != nil {
			return err
		}

		for i, offset := int32(0), toc.Offset; i < toc.Count; i, offset = i+1, offset+itemsize {
			if err := writeItem(offset); err != nil {
				return err
			}
		}

		if _, err := fmt.Fprintln(w); err != nil {
			return err
		}

		toff += TocLength
	}

	return nil
}

// Write creates a writable representation of a MMV dump
// and writes it to the passed writer.
func Write(
	w io.Writer,
	header *Header,
	tocs []*Toc,
	metrics map[uint64]Metric,
	values map[uint64]*Value,
	instances map[uint64]Instance,
	indoms map[uint64]*InstanceDomain,
	strings map[uint64]*String,
) error {
	if _, err := fmt.Fprintf(w, `Version   = %v
Generated = %v
Toc Count = %v
Cluster   = %v
Process   = %v
Flags     = 0x%x

`, header.Version, header.G1, header.Toc, header.Cluster, header.Process, int(header.Flag)); err != nil {
		return err
	}

	return writeComponents(w, header, tocs, metrics, values, instances, indoms, strings)
}
