package speed

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/performancecopilot/speed/bytewriter"
)

// byte lengths of different components in an mmv file
const (
	HeaderLength         = 40
	TocLength            = 16
	Metric1Length        = 104
	Metric2Length        = 48
	ValueLength          = 32
	Instance1Length      = 80
	Instance2Length      = 24
	InstanceDomainLength = 32
	StringLength         = 256
)

// MaxV1NameLength is the maximum length for a metric/instance name
// under MMV format 1
const MaxV1NameLength = 63

// MaxDataValueSize is the maximum byte length for a stored metric value, unless it is a string
const MaxDataValueSize = 16

// EraseFileOnStop if set to true, will also delete the memory mapped file
var EraseFileOnStop = false

// Client defines the interface for a type that can talk to an instrumentation agent
type Client interface {
	// a client must contain a registry of metrics
	Registry() Registry

	// starts monitoring
	Start() error

	// Start that will panic on failure
	MustStart()

	// stop monitoring
	Stop() error

	// Stop that will panic on failure
	MustStop()

	// adds a metric to be monitored
	Register(Metric) error

	// tries to add a metric to be written and panics on error
	MustRegister(Metric)

	// adds metric from a string
	RegisterString(string, interface{}, MetricType, MetricSemantics, MetricUnit) (Metric, error)

	// tries to add a metric from a string and panics on an error
	MustRegisterString(string, interface{}, MetricType, MetricSemantics, MetricUnit) Metric
}

///////////////////////////////////////////////////////////////////////////////

func mmvFileLocation(name string) (string, error) {
	if strings.ContainsRune(name, os.PathSeparator) {
		return "", errors.New("name cannot have path separator")
	}

	tdir, present := config["PCP_TMP_DIR"]
	var loc string
	if present {
		loc = filepath.Join(rootPath, tdir)
	} else {
		loc = os.TempDir()
	}

	return filepath.Join(loc, "mmv", name), nil
}

// PCPClusterIDBitLength is the bit length of the cluster id
// for a set of PCP metrics
const PCPClusterIDBitLength = 12

// MMVFlag represents an enumerated type to represent mmv flag values
type MMVFlag int

// values for MMVFlag
const (
	NoPrefixFlag MMVFlag = 1 << iota
	ProcessFlag
	SentinelFlag
)

//go:generate stringer -type=MMVFlag

// PCPClient implements a client that can generate instrumentation for PCP
type PCPClient struct {
	mutex sync.Mutex

	loc       string  // absolute location of the mmv file
	clusterID uint32  // cluster identifier for the writer
	flag      MMVFlag // write flag

	r *PCPRegistry // current registry

	writer bytewriter.Writer

	instanceoffsetc chan int
	indomoffsetc    chan int
	metricoffsetc   chan int
	valueoffsetc    chan int
	stringoffsetc   chan int
}

// NewPCPClient initializes a new PCPClient object
func NewPCPClient(name string) (*PCPClient, error) {
	return NewPCPClientWithRegistry(name, NewPCPRegistry())
}

// NewPCPClientWithRegistry initializes a new PCPClient object with the given registry
func NewPCPClientWithRegistry(name string, registry *PCPRegistry) (*PCPClient, error) {
	fileLocation, err := mmvFileLocation(name)
	if err != nil {
		return nil, errors.Wrap(err, "could not get a location for storing MMV file")
	}

	return &PCPClient{
		loc:       fileLocation,
		r:         registry,
		clusterID: hash(name, PCPClusterIDBitLength),
		flag:      ProcessFlag,
	}, nil
}

// Registry returns a writer's registry
func (c *PCPClient) Registry() Registry {
	return c.r
}

// SetFlag sets the MMVflag for the client
func (c *PCPClient) SetFlag(flag MMVFlag) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.r.mapped {
		return errors.New("cannot set mmv flag for an active client")
	}

	c.flag = flag
	return nil
}

func (c *PCPClient) tocCount() int {
	ans := 2

	if c.r.InstanceCount() > 0 {
		ans += 2
	}

	if c.r.StringCount() > 0 {
		ans++
	}

	return ans
}

// Length returns the byte length of data in the mmv file written by the current writer
func (c *PCPClient) Length() int {
	var (
		InstanceLength = Instance1Length
		MetricLength   = Metric1Length
	)

	if c.r.version2 {
		InstanceLength = Instance2Length
		MetricLength = Metric2Length
	}

	return HeaderLength +
		(c.tocCount() * TocLength) +
		(c.r.InstanceCount() * InstanceLength) +
		(c.r.InstanceDomainCount() * InstanceDomainLength) +
		(c.r.MetricCount() * MetricLength) +
		(c.r.ValuesCount() * ValueLength) +
		(c.r.StringCount() * StringLength)
}

// Start dumps existing registry data
func (c *PCPClient) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	l := c.Length()

	writer, err := bytewriter.NewMemoryMappedWriter(c.loc, l)
	if err != nil {
		return errors.Wrap(err, "cannot create MemoryMappedBuffer in client")
	}

	c.writer = writer
	c.start()
	c.r.mapped = true
	return nil
}

func (c *PCPClient) start() {
	var (
		InstanceLength = Instance1Length
		MetricLength   = Metric1Length
	)

	if c.r.version2 {
		InstanceLength = Instance2Length
		MetricLength = Metric2Length
	}

	c.r.indomoffset = HeaderLength + TocLength*c.tocCount()
	c.r.instanceoffset = c.r.indomoffset + InstanceDomainLength*c.r.InstanceDomainCount()
	c.r.metricsoffset = c.r.instanceoffset + InstanceLength*c.r.InstanceCount()
	c.r.valuesoffset = c.r.metricsoffset + MetricLength*c.r.MetricCount()
	c.r.stringsoffset = c.r.valuesoffset + ValueLength*c.r.ValuesCount()

	if c.r.InstanceDomainCount() > 0 {
		c.instanceoffsetc, c.indomoffsetc = make(chan int, 1), make(chan int, 1)

		c.instanceoffsetc <- c.r.instanceoffset
		c.indomoffsetc <- c.r.indomoffset
	}

	if c.r.MetricCount() > 0 {
		c.metricoffsetc, c.valueoffsetc = make(chan int, 1), make(chan int, 1)

		c.metricoffsetc <- c.r.metricsoffset
		c.valueoffsetc <- c.r.valuesoffset
	}

	if c.r.StringCount() > 0 {
		c.stringoffsetc = make(chan int, 1)
		c.stringoffsetc <- c.r.stringsoffset
	}

	genc, g2offc := make(chan int64), make(chan int)

	go c.writeHeaderBlock(genc, g2offc)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		c.writeTocBlock()
		wg.Done()
	}()

	go func() {
		// instance domains **have** to be written before metrics
		// as metrics need instance offsets and multiple metrics
		// can have the same indom, so they need to be cached
		c.writeInstanceDomains()
		c.writeMetrics()
		wg.Done()
	}()

	gen, g2off := <-genc, <-g2offc
	wg.Wait()

	// must *always* be the last thing to happen
	_ = c.writer.MustWriteInt64(gen, g2off)
}

func (c *PCPClient) writeHeaderBlock(genc chan int64, g2offc chan int) {
	// tag
	c.writer.MustWriteString("MMV", 0)

	var pos int

	// version
	if c.r.version2 {
		pos = c.writer.MustWriteUint32(2, 4)
	} else {
		pos = c.writer.MustWriteUint32(1, 4)
	}

	// generation
	gen := time.Now().Unix()
	pos = c.writer.MustWriteInt64(gen, pos)

	g2off := pos
	pos = c.writer.MustWriteInt64(0, pos)

	// tocCount
	pos = c.writer.MustWriteInt32(int32(c.tocCount()), pos)

	// flag mask
	pos = c.writer.MustWriteInt32(int32(c.flag), pos)

	// process identifier
	pos = c.writer.MustWriteInt32(int32(os.Getpid()), pos)

	// cluster identifier
	_ = c.writer.MustWriteUint32(c.clusterID, pos)

	// NOTE: the order here is important, should be same as in start()
	// or deadlock
	genc <- gen
	g2offc <- g2off

	return
}

func (c *PCPClient) writeTocBlock() {
	var wg sync.WaitGroup
	tocpos := HeaderLength

	wg.Add(c.tocCount())

	// instance domains toc
	if c.r.InstanceDomainCount() > 0 {
		go func(pos int) {
			// 1 is the identifier for instance domains
			c.writeSingleToc(pos, 1, c.r.InstanceDomainCount(), c.r.indomoffset)
			wg.Done()
		}(tocpos)
		tocpos += TocLength
	}

	// instances toc
	if c.r.InstanceCount() > 0 {
		go func(pos int) {
			// 2 is the identifier for instances
			c.writeSingleToc(pos, 2, c.r.InstanceCount(), c.r.instanceoffset)
			wg.Done()
		}(tocpos)
		tocpos += TocLength
	}

	// metrics and values toc
	metricsoffset, valuesoffset := c.r.metricsoffset, c.r.valuesoffset
	if c.r.MetricCount() == 0 {
		metricsoffset, valuesoffset = 0, 0
	}

	go func(pos int) {
		// 3 is the identifier for metrics
		c.writeSingleToc(pos, 3, c.r.MetricCount(), metricsoffset)
		wg.Done()
	}(tocpos)
	tocpos += TocLength

	go func(pos int) {
		// 4 is the identifier for values
		c.writeSingleToc(pos, 4, c.r.ValuesCount(), valuesoffset)
		wg.Done()
	}(tocpos)
	tocpos += TocLength

	// strings toc
	if c.r.StringCount() > 0 {
		go func(pos int) {
			// 5 is the identifier for strings
			c.writeSingleToc(pos, 5, c.r.StringCount(), c.r.stringsoffset)
			wg.Done()
		}(tocpos)
	}

	wg.Wait()
}

func (c *PCPClient) writeSingleToc(pos, identifier, count, offset int) {
	pos = c.writer.MustWriteInt32(int32(identifier), pos)
	pos = c.writer.MustWriteInt32(int32(count), pos)
	_ = c.writer.MustWriteUint64(uint64(offset), pos)
}

func (c *PCPClient) writeInstanceDomains() {
	var wg sync.WaitGroup
	wg.Add(c.r.InstanceDomainCount())

	for _, indom := range c.r.instanceDomains {
		go func(indom *PCPInstanceDomain) {
			c.writeInstanceDomain(indom)
			wg.Done()
		}(indom)
	}

	wg.Wait()
}

func (c *PCPClient) writeInstanceDomain(indom *PCPInstanceDomain) {
	off := <-c.indomoffsetc
	c.indomoffsetc <- off + InstanceDomainLength

	InstanceLength := Instance1Length
	if c.r.version2 {
		InstanceLength = Instance2Length
	}

	inoff := off
	ioff := <-c.instanceoffsetc
	c.instanceoffsetc <- ioff + InstanceLength*indom.InstanceCount()

	var wg sync.WaitGroup
	wg.Add(indom.InstanceCount())

	off = c.writer.MustWriteUint32(indom.id, off)
	off = c.writer.MustWriteInt32(int32(indom.InstanceCount()), off)
	off = c.writer.MustWriteInt64(int64(ioff), off)

	for _, i := range indom.instances {
		go func(i *pcpInstance, offset int) {
			c.writeInstance(i, inoff, offset)
			wg.Done()
		}(i, ioff)
		ioff += InstanceLength
	}

	so, lo := 0, 0

	if indom.shortDescription != "" {
		so = <-c.stringoffsetc
		c.stringoffsetc <- so + StringLength

		c.writer.MustWriteString(indom.shortDescription, so)
	}

	if indom.longDescription != "" {
		lo = <-c.stringoffsetc
		c.stringoffsetc <- lo + StringLength

		c.writer.MustWriteString(indom.longDescription, lo)
	}

	off = c.writer.MustWriteUint64(uint64(so), off)
	_ = c.writer.MustWriteUint64(uint64(lo), off)

	wg.Wait()
}

func (c *PCPClient) writeInstance(i *pcpInstance, indomoff int, off int) {
	i.offset = off

	off = c.writer.MustWriteInt64(int64(indomoff), off)
	off = c.writer.MustWriteInt32(0, off)
	off = c.writer.MustWriteUint32(i.id, off)

	if c.r.version2 {
		soff := <-c.stringoffsetc
		c.stringoffsetc <- soff + StringLength

		c.writer.MustWriteUint64(uint64(soff), off)
		c.writer.MustWriteString(i.name, soff)
	} else {
		c.writer.MustWriteString(i.name, off)
	}
}

func (c *PCPClient) writeMetrics() {
	var wg sync.WaitGroup

	launchSingletonMetric := func(metric *pcpSingletonMetric) {
		go func() {
			c.writeSingletonMetric(metric)
			wg.Done()
		}()
	}

	launchInstanceMetric := func(metric *pcpInstanceMetric) {
		go func() {
			c.writeInstanceMetric(metric)
			wg.Done()
		}()
	}

	wg.Add(c.r.MetricCount())
	for _, m := range c.r.metrics {
		switch metric := m.(type) {
		case *PCPSingletonMetric:
			launchSingletonMetric(metric.pcpSingletonMetric)
		case *PCPCounter:
			launchSingletonMetric(metric.pcpSingletonMetric)
		case *PCPGauge:
			launchSingletonMetric(metric.pcpSingletonMetric)
		case *PCPTimer:
			launchSingletonMetric(metric.pcpSingletonMetric)
		case *PCPInstanceMetric:
			launchInstanceMetric(metric.pcpInstanceMetric)
		case *PCPCounterVector:
			launchInstanceMetric(metric.pcpInstanceMetric)
		case *PCPGaugeVector:
			launchInstanceMetric(metric.pcpInstanceMetric)
		case *PCPHistogram:
			launchInstanceMetric(metric.pcpInstanceMetric)
		}
	}

	wg.Wait()
}

func (c *PCPClient) writeSingletonMetric(m *pcpSingletonMetric) {
	var wg sync.WaitGroup
	wg.Add(2)

	doff := <-c.metricoffsetc

	go func() {
		c.writeMetricDesc(m.pcpMetricDesc, m.Indom(), doff)
		wg.Done()
	}()

	off := <-c.valueoffsetc
	c.valueoffsetc <- off + ValueLength

	go func(offset int) {
		m.update = c.writeValue(m.t, m.val, offset)
		wg.Done()
	}(off)

	off = c.writer.MustWriteInt64(int64(doff), off+MaxDataValueSize)
	_ = c.writer.MustWriteInt64(0, off)

	wg.Wait()
}

func (c *PCPClient) writeInstanceMetric(m *pcpInstanceMetric) {
	var wg sync.WaitGroup
	wg.Add(1 + m.Indom().InstanceCount())

	doff := <-c.metricoffsetc

	go func() {
		c.writeMetricDesc(m.pcpMetricDesc, m.Indom(), doff)
		wg.Done()
	}()

	for name, i := range m.indom.instances {
		off := <-c.valueoffsetc
		c.valueoffsetc <- off + ValueLength

		go func(i *instanceValue, offset int) {
			i.update = c.writeValue(m.t, i.val, offset)
			wg.Done()
		}(m.vals[name], off)

		off = c.writer.MustWriteInt64(int64(doff), off+MaxDataValueSize)
		_ = c.writer.MustWriteInt64(int64(i.offset), off)
	}

	wg.Wait()
}

func (c *PCPClient) writeMetricDesc(desc *pcpMetricDesc, indom *PCPInstanceDomain, off int) {
	if c.r.version2 {
		c.metricoffsetc <- off + Metric2Length

		noff := <-c.stringoffsetc
		c.stringoffsetc <- noff + StringLength

		off = c.writer.MustWriteUint64(uint64(noff), off)
		c.writer.MustWriteString(desc.name, noff)
	} else {
		c.metricoffsetc <- off + Metric1Length

		c.writer.MustWriteString(desc.name, off)
		off += MaxV1NameLength + 1
	}

	off = c.writer.MustWriteUint32(desc.id, off)
	off = c.writer.MustWriteInt32(int32(desc.t), off)
	off = c.writer.MustWriteInt32(int32(desc.sem), off)
	off = c.writer.MustWriteUint32(desc.u.PMAPI(), off)

	if indom != nil {
		off = c.writer.MustWriteUint32(indom.ID(), off)
	} else {
		off = c.writer.MustWriteInt32(-1, off)
	}

	off = c.writer.MustWriteInt32(0, off)

	so, lo := 0, 0

	if desc.shortDescription != "" {
		so = <-c.stringoffsetc
		c.stringoffsetc <- so + StringLength

		c.writer.MustWriteString(desc.shortDescription, so)
	}

	if desc.longDescription != "" {
		lo = <-c.stringoffsetc
		c.stringoffsetc <- lo + StringLength

		c.writer.MustWriteString(desc.longDescription, lo)
	}

	off = c.writer.MustWriteUint64(uint64(so), off)
	_ = c.writer.MustWriteUint64(uint64(lo), off)
}

func (c *PCPClient) writeValue(t MetricType, val interface{}, offset int) updateClosure {
	if t == StringType {
		pos := c.writer.MustWriteUint64(StringLength-1, offset)

		offset = <-c.stringoffsetc
		c.stringoffsetc <- offset + StringLength

		c.writer.MustWriteUint64(uint64(offset), pos)
	}

	update := newupdateClosure(offset, c.writer)
	_ = update(val)

	return update
}

// MustStart is a start that panics
func (c *PCPClient) MustStart() {
	if err := c.Start(); err != nil {
		panic(err)
	}
}

// Stop removes existing mapping and cleans up
func (c *PCPClient) Stop() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.r.mapped {
		return errors.New("trying to stop an already stopped mapping")
	}

	c.stop()

	c.r.mapped = false

	err := c.writer.(*bytewriter.MemoryMappedWriter).Unmap(EraseFileOnStop)
	c.writer = nil
	if err != nil {
		return errors.Wrap(err, "client: error unmapping MemoryMappedBuffer")
	}

	return nil
}

func (c *PCPClient) stop() {
	c.instanceoffsetc, c.indomoffsetc = nil, nil
	c.metricoffsetc, c.valueoffsetc = nil, nil
	c.stringoffsetc = nil
}

// MustStop is a stop that panics
func (c *PCPClient) MustStop() {
	if err := c.Stop(); err != nil {
		panic(err)
	}
}

// Register is simply a shorthand for Registry().AddMetric
func (c *PCPClient) Register(m Metric) error { return c.r.AddMetric(m) }

// MustRegister is simply a Register that can panic
func (c *PCPClient) MustRegister(m Metric) {
	if err := c.Register(m); err != nil {
		panic(err)
	}
}

// RegisterIndom is simply a shorthand for Registry().AddInstanceDomain
func (c *PCPClient) RegisterIndom(indom InstanceDomain) error {
	return c.r.AddInstanceDomain(indom)
}

// MustRegisterIndom is simply a RegisterIndom that can panic
func (c *PCPClient) MustRegisterIndom(indom InstanceDomain) {
	if err := c.RegisterIndom(indom); err != nil {
		panic(err)
	}
}

// RegisterString is simply a shorthand for Registry().AddMetricByString
func (c *PCPClient) RegisterString(str string, val interface{}, t MetricType, s MetricSemantics, u MetricUnit) (Metric, error) {
	return c.r.AddMetricByString(str, val, t, s, u)
}

// MustRegisterString is simply a RegisterString that panics
func (c *PCPClient) MustRegisterString(str string, val interface{}, t MetricType, s MetricSemantics, u MetricUnit) Metric {
	if m, err := c.RegisterString(str, val, t, s, u); err != nil {
		panic(err)
	} else {
		return m
	}
}
