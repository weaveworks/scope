package speed

import (
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	histogram "github.com/codahale/hdrhistogram"
	"github.com/pkg/errors"

	"github.com/performancecopilot/speed/bytewriter"
)

// MetricType is an enumerated type representing all valid types for a metric.
type MetricType int32

// Possible values for a MetricType.
const (
	Int32Type MetricType = iota
	Uint32Type
	Int64Type
	Uint64Type
	FloatType
	DoubleType
	StringType
)

//go:generate stringer -type=MetricType

func (m MetricType) isCompatibleInt(val int) bool {
	v := int64(val)
	switch m {
	case Int32Type:
		return v >= math.MinInt32 && v <= math.MaxInt32
	case Int64Type:
		return v >= math.MinInt64 && v <= math.MaxInt64
	case Uint32Type:
		return v >= 0 && v <= math.MaxUint32
	case Uint64Type:
		return v >= 0 && uint64(v) <= math.MaxUint64
	}
	return false
}

func (m MetricType) isCompatibleUint(val uint) bool {
	switch {
	case val <= math.MaxUint32:
		return m == Uint32Type || m == Uint64Type
	default:
		return m == Uint64Type
	}
}

func (m MetricType) isCompatibleFloat(val float64) bool {
	switch {
	case val >= -math.MaxFloat32 && val <= math.MaxFloat32:
		return m == FloatType || m == DoubleType
	default:
		return m == DoubleType
	}
}

// IsCompatible checks if the passed value is compatible with the current MetricType.
func (m MetricType) IsCompatible(val interface{}) bool {
	switch v := val.(type) {
	case int:
		return m.isCompatibleInt(v)
	case int32:
		return m == Int32Type
	case int64:
		return m == Int64Type
	case uint:
		return m.isCompatibleUint(v)
	case uint32:
		return m == Uint32Type
	case uint64:
		return m == Uint64Type
	case float32:
		return m == FloatType
	case float64:
		return m.isCompatibleFloat(v)
	case string:
		return m == StringType
	}
	return false
}

// resolveInt will resolve an int to one of the 4 compatible types.
func (m MetricType) resolveInt(val interface{}) interface{} {
	if vi, isInt := val.(int); isInt {
		switch m {
		case Int64Type:
			return int64(vi)
		case Uint32Type:
			return uint32(vi)
		case Uint64Type:
			return uint64(vi)
		}
		return int32(val.(int))
	}

	if vui, isUint := val.(uint); isUint {
		if m == Uint64Type {
			return uint64(vui)
		}
		return uint32(vui)
	}

	return val
}

// resolveFloat will resolve a float64 to one of the 2 compatible types.
func (m MetricType) resolveFloat(val interface{}) interface{} {
	_, isFloat64 := val.(float64)
	if isFloat64 && m == FloatType {
		return float32(val.(float64))
	}

	return val
}

func (m MetricType) resolve(val interface{}) interface{} {
	val = m.resolveInt(val)
	val = m.resolveFloat(val)

	return val
}

///////////////////////////////////////////////////////////////////////////////

// MetricUnit defines the interface for a unit type for speed.
type MetricUnit interface {
	fmt.Stringer

	// return 32 bit PMAPI representation for the unit
	// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/pmapi.h#L61-L101
	PMAPI() uint32

	// add a space unit to the current unit at a specific dimension
	Space(SpaceUnit, int8) MetricUnit

	// add a time unit to the current unit at a specific dimension
	Time(TimeUnit, int8) MetricUnit

	// add a count unit to the current unit at a specific dimension
	Count(CountUnit, int8) MetricUnit
}

// internal struct for supporting composite units,
// based on the implementation inside hornet
//
// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#327
type metricUnit struct {
	repr uint32
}

// NewMetricUnit returns a new object for initialization
func NewMetricUnit() MetricUnit {
	return &metricUnit{}
}

func (m *metricUnit) PMAPI() uint32 { return m.repr }

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#375
func (m *metricUnit) Space(s SpaceUnit, dimension int8) MetricUnit {
	if dimension < -8 || dimension > 7 {
		panic("dimension has to be between -8 and 7 inclusive")
	}

	m.repr |= uint32(s)
	m.repr |= (uint32(dimension) & 0xF) << 28
	return m
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#383
func (m *metricUnit) Time(t TimeUnit, dimension int8) MetricUnit {
	if dimension < -8 || dimension > 7 {
		panic("dimension has to be between -8 and 7 inclusive")
	}

	m.repr |= uint32(t)
	m.repr |= (uint32(dimension) & 0xF) << 24
	return m
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#391
func (m *metricUnit) Count(c CountUnit, dimension int8) MetricUnit {
	if dimension < -8 || dimension > 7 {
		panic("dimension has to be between -8 and 7 inclusive")
	}

	m.repr |= uint32(c)
	m.repr |= (uint32(dimension) & 0xF) << 20
	return m
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#410
func (m *metricUnit) SpaceDim() int8 {
	return int8(int32(m.repr) >> 28)
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#398
func (m *metricUnit) SpaceScale() SpaceUnit {
	d := m.SpaceDim()
	if d == 0 {
		panic("no space scale on unit")
	}

	return SpaceUnit(1<<28 | (uint32((m.repr>>16)&0xF))<<16)
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#410
func (m *metricUnit) TimeDim() int8 {
	return int8(int32(m.repr<<4) >> 28)
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#398
func (m *metricUnit) TimeScale() TimeUnit {
	d := m.TimeDim()
	if d == 0 {
		panic("no time scale on unit")
	}

	return TimeUnit(1<<24 | (uint32((m.repr>>12)&0xF))<<12)
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#410
func (m *metricUnit) CountDim() int8 {
	return int8(int32(m.repr<<8) >> 28)
}

// https://docs.rs/hornet/0.1.0/src/hornet/client/metric/mod.rs.html#398
func (m *metricUnit) CountScale() CountUnit {
	d := m.CountDim()
	if d == 0 {
		panic("no count scale on unit")
	}

	return CountUnit(1<<20 | (uint32((m.repr>>8)&0xF))>>8)
}

func (m *metricUnit) String() string {
	sd, td, cd := m.SpaceDim(), m.TimeDim(), m.CountDim()

	ans := ""

	if sd != 0 {
		ans = ans + m.SpaceScale().String() + "^" + strconv.Itoa(int(m.SpaceDim()))
	}

	if td != 0 {
		ans = ans + m.TimeScale().String() + "^" + strconv.Itoa(int(m.TimeDim()))
	}

	if cd != 0 {
		ans = ans + m.CountScale().String() + "^" + strconv.Itoa(int(m.CountDim()))
	}

	return ans
}

// SpaceUnit is an enumerated type representing all units for space.
type SpaceUnit uint32

// Possible values for SpaceUnit.
const (
	ByteUnit SpaceUnit = 1<<28 | iota<<16
	KilobyteUnit
	MegabyteUnit
	GigabyteUnit
	TerabyteUnit
	PetabyteUnit
	ExabyteUnit
)

//go:generate stringer -type=SpaceUnit

// PMAPI returns the PMAPI representation for a SpaceUnit
// for space units bits 0-3 are 1 and bits 13-16 are scale
func (s SpaceUnit) PMAPI() uint32 {
	return uint32(s)
}

// Space adds a space unit to the current unit at a specific dimension
func (s SpaceUnit) Space(SpaceUnit, int8) MetricUnit {
	panic("Cannot add another space unit")
}

// Time adds a time unit to the current unit at a specific dimension
func (s SpaceUnit) Time(t TimeUnit, dimension int8) MetricUnit {
	return (&metricUnit{uint32(s)}).Time(t, dimension)
}

// Count adds a count unit to the current unit at a specific dimension
func (s SpaceUnit) Count(c CountUnit, dimension int8) MetricUnit {
	return (&metricUnit{uint32(s)}).Count(c, dimension)
}

// TimeUnit is an enumerated type representing all possible units for representing time.
type TimeUnit uint32

// Possible Values for TimeUnit.
// for time units bits 4-7 are 1 and bits 17-20 are scale.
const (
	NanosecondUnit TimeUnit = 1<<24 | iota<<12
	MicrosecondUnit
	MillisecondUnit
	SecondUnit
	MinuteUnit
	HourUnit
)

//go:generate stringer -type=TimeUnit

// PMAPI returns the PMAPI representation for a TimeUnit.
func (t TimeUnit) PMAPI() uint32 {
	return uint32(t)
}

// Space adds a space unit to the current unit at a specific dimension
func (t TimeUnit) Space(s SpaceUnit, dimension int8) MetricUnit {
	return (&metricUnit{uint32(t)}).Space(s, dimension)
}

// Time adds a time unit to the current unit at a specific dimension
func (t TimeUnit) Time(TimeUnit, int8) MetricUnit {
	panic("Cannot add another time unit")
}

// Count adds a count unit to the current unit at a specific dimension
func (t TimeUnit) Count(c CountUnit, dimension int8) MetricUnit {
	return (&metricUnit{uint32(t)}).Count(c, dimension)
}

// CountUnit is a type representing a counted quantity.
type CountUnit uint32

// OneUnit represents the only CountUnit.
// For count units bits 8-11 are 1 and bits 21-24 are scale.
const OneUnit CountUnit = 1<<20 | iota<<8

//go:generate stringer -type=CountUnit

// PMAPI returns the PMAPI representation for a CountUnit.
func (c CountUnit) PMAPI() uint32 {
	return uint32(c)
}

// Space adds a space unit to the current unit at a specific dimension
func (c CountUnit) Space(s SpaceUnit, dimension int8) MetricUnit {
	return (&metricUnit{uint32(c)}).Space(s, dimension)
}

// Time adds a time unit to the current unit at a specific dimension
func (c CountUnit) Time(t TimeUnit, dimension int8) MetricUnit {
	return (&metricUnit{uint32(c)}).Time(t, dimension)
}

// Count adds a count unit to the current unit at a specific dimension
func (c CountUnit) Count(CountUnit, int8) MetricUnit {
	panic("Cannot add another time unit")
}

///////////////////////////////////////////////////////////////////////////////

// MetricSemantics represents an enumerated type representing the possible
// values for the semantics of a metric.
type MetricSemantics int32

// Possible values for MetricSemantics.
const (
	NoSemantics MetricSemantics = iota
	CounterSemantics
	_
	InstantSemantics
	DiscreteSemantics
)

//go:generate stringer -type=MetricSemantics

///////////////////////////////////////////////////////////////////////////////

// Metric defines the general interface a type needs to implement to qualify
// as a valid PCP metric.
type Metric interface {
	// gets the unique id generated for this metric
	ID() uint32

	// gets the name for the metric
	Name() string

	// gets the type of a metric
	Type() MetricType

	// gets the unit of a metric
	Unit() MetricUnit

	// gets the semantics for a metric
	Semantics() MetricSemantics

	// gets the description of a metric
	Description() string
}

///////////////////////////////////////////////////////////////////////////////

// SingletonMetric defines the interface for a metric that stores only one value.
type SingletonMetric interface {
	Metric

	// gets the value of the metric
	Val() interface{}

	// sets the value of the metric to a value, optionally returns an error on failure
	Set(interface{}) error

	// tries to set and panics on error
	MustSet(interface{})
}

///////////////////////////////////////////////////////////////////////////////

// InstanceMetric defines the interface for a metric that stores multiple values
// in instances and instance domains.
type InstanceMetric interface {
	Metric

	// gets the value of a particular instance
	ValInstance(string) (interface{}, error)

	// sets the value of a particular instance
	SetInstance(interface{}, string) error

	// tries to set the value of a particular instance and panics on error
	MustSetInstance(interface{}, string)

	// returns a slice containing all instances in the metric
	Instances() []string
}

///////////////////////////////////////////////////////////////////////////////

// PCPMetric defines the interface for a metric that is compatible with PCP.
type PCPMetric interface {
	Metric

	// a PCPMetric will always have an instance domain, even if it is nil
	Indom() *PCPInstanceDomain

	ShortDescription() string

	LongDescription() string
}

///////////////////////////////////////////////////////////////////////////////

// PCPMetricItemBitLength is the maximum bit size of a PCP Metric id.
//
// see: https://github.com/performancecopilot/pcp/blob/master/src/include/pcp/impl.h#L102-L121
const PCPMetricItemBitLength = 10

// pcpMetricDesc is a metric metadata wrapper
// each metric type can wrap its metadata by containing a pcpMetricDesc type and
// only define its own specific properties assuming pcpMetricDesc will handle the rest.
//
// when writing, this type is supposed to map directly to the pmDesc struct as defined in PCP core.
type pcpMetricDesc struct {
	id                                uint32          // unique metric id
	name                              string          // the name
	t                                 MetricType      // the type of a metric
	sem                               MetricSemantics // the semantics
	u                                 MetricUnit      // the unit
	shortDescription, longDescription string
}

// newpcpMetricDesc creates a new Metric Description wrapper type.
func newpcpMetricDesc(n string, t MetricType, s MetricSemantics, u MetricUnit, desc ...string) (*pcpMetricDesc, error) {
	if n == "" {
		return nil, errors.New("Metric name cannot be empty")
	}

	if len(n) > StringLength {
		return nil, errors.New("metric name is too long")
	}

	if len(desc) > 2 {
		return nil, errors.New("only 2 optional strings allowed, short and long descriptions")
	}

	shortdesc, longdesc := "", ""

	if len(desc) > 0 {
		shortdesc = desc[0]
	}

	if len(desc) > 1 {
		longdesc = desc[1]
	}

	return &pcpMetricDesc{
		hash(n, PCPMetricItemBitLength),
		n, t, s, u,
		shortdesc, longdesc,
	}, nil
}

// ID returns the generated id for PCPMetric.
func (md *pcpMetricDesc) ID() uint32 { return md.id }

// Name returns the generated id for PCPMetric.
func (md *pcpMetricDesc) Name() string {
	return md.name
}

// Semantics returns the current stored value for PCPMetric.
func (md *pcpMetricDesc) Semantics() MetricSemantics { return md.sem }

// Unit returns the unit for PCPMetric.
func (md *pcpMetricDesc) Unit() MetricUnit { return md.u }

// Type returns the type for PCPMetric.
func (md *pcpMetricDesc) Type() MetricType { return md.t }

// ShortDescription returns the shortdesc value.
func (md *pcpMetricDesc) ShortDescription() string { return md.shortDescription }

// LongDescription returns the longdesc value.
func (md *pcpMetricDesc) LongDescription() string { return md.longDescription }

// Description returns the description for PCPMetric.
func (md *pcpMetricDesc) Description() string {
	return md.shortDescription + "\n" + md.longDescription
}

///////////////////////////////////////////////////////////////////////////////

// updateClosure is a closure that will write the modified value of a metric on disk.
type updateClosure func(interface{}) error

// newupdateClosure creates a new update closure for an offset, type and buffer.
func newupdateClosure(offset int, writer bytewriter.Writer) updateClosure {
	return func(val interface{}) error {
		if _, isString := val.(string); isString {
			writer.MustWrite(make([]byte, StringLength), offset)
		}

		_, err := writer.WriteVal(val, offset)
		return err
	}
}

///////////////////////////////////////////////////////////////////////////////

// pcpSingletonMetric defines an embeddable base singleton metric.
type pcpSingletonMetric struct {
	*pcpMetricDesc
	val    interface{}
	update updateClosure
}

// newpcpSingletonMetric creates a new instance of pcpSingletonMetric.
func newpcpSingletonMetric(val interface{}, desc *pcpMetricDesc) (*pcpSingletonMetric, error) {
	if !desc.t.IsCompatible(val) {
		return nil, errors.Errorf("type %v is not compatible with value %v(%T)", desc.t, val, val)
	}

	val = desc.t.resolve(val)
	return &pcpSingletonMetric{desc, val, nil}, nil
}

// set Sets the current value of pcpSingletonMetric.
func (m *pcpSingletonMetric) set(val interface{}) error {
	if !m.t.IsCompatible(val) {
		return errors.Errorf("value %v is incompatible with MetricType %v", val, m.t)
	}

	val = m.t.resolve(val)

	if val != m.val {
		if m.update != nil {
			err := m.update(val)
			if err != nil {
				return err
			}
		}
		m.val = val
	}

	return nil
}

func (m *pcpSingletonMetric) Indom() *PCPInstanceDomain { return nil }

///////////////////////////////////////////////////////////////////////////////

// PCPSingletonMetric defines a singleton metric with no instance domain
// only a value and a valueoffset.
type PCPSingletonMetric struct {
	*pcpSingletonMetric
	mutex sync.RWMutex
}

// NewPCPSingletonMetric creates a new instance of PCPSingletonMetric
// it takes 2 extra optional strings as short and long description parameters,
// which on not being present are set to blank strings.
func NewPCPSingletonMetric(val interface{}, name string, t MetricType, s MetricSemantics, u MetricUnit, desc ...string) (*PCPSingletonMetric, error) {
	d, err := newpcpMetricDesc(name, t, s, u, desc...)
	if err != nil {
		return nil, err
	}

	sm, err := newpcpSingletonMetric(val, d)
	if err != nil {
		return nil, err
	}

	return &PCPSingletonMetric{sm, sync.RWMutex{}}, nil
}

// Val returns the current Set value of PCPSingletonMetric.
func (m *PCPSingletonMetric) Val() interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.val
}

// Set Sets the current value of PCPSingletonMetric.
func (m *PCPSingletonMetric) Set(val interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.set(val)
}

// MustSet is a Set that panics on failure.
func (m *PCPSingletonMetric) MustSet(val interface{}) {
	if err := m.Set(val); err != nil {
		panic(err)
	}
}

func (m *PCPSingletonMetric) String() string {
	return fmt.Sprintf("Val: %v\n%v", m.val, m.Description())
}

///////////////////////////////////////////////////////////////////////////////

// Counter defines a metric that holds a single value that can only be incremented.
type Counter interface {
	Metric

	Val() int64
	Set(int64) error

	Inc(int64) error
	MustInc(int64)

	Up() // same as MustInc(1)
}

///////////////////////////////////////////////////////////////////////////////

// PCPCounter implements a PCP compatible Counter Metric.
type PCPCounter struct {
	*pcpSingletonMetric
	mutex sync.RWMutex
}

// NewPCPCounter creates a new PCPCounter instance.
// It requires an initial int64 value and a metric name for construction.
// optionally it can also take a couple of description strings that are used as
// short and long descriptions respectively.
// Internally it creates a PCP SingletonMetric with Int64Type, CounterSemantics
// and CountUnit.
func NewPCPCounter(val int64, name string, desc ...string) (*PCPCounter, error) {
	d, err := newpcpMetricDesc(name, Int64Type, CounterSemantics, OneUnit, desc...)
	if err != nil {
		return nil, err
	}

	sm, err := newpcpSingletonMetric(val, d)
	if err != nil {
		return nil, err
	}

	return &PCPCounter{sm, sync.RWMutex{}}, nil
}

// Val returns the current value of the counter.
func (c *PCPCounter) Val() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.val.(int64)
}

// Set sets the value of the counter.
func (c *PCPCounter) Set(val int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	v := c.val.(int64)

	if val < v {
		return errors.Errorf("cannot set counter to %v, current value is %v and PCP counters cannot go backwards", val, v)
	}

	return c.set(val)
}

// Inc increases the stored counter's value by the passed increment.
func (c *PCPCounter) Inc(val int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if val < 0 {
		return errors.New("cannot decrement a counter")
	}

	if val == 0 {
		return nil
	}

	v := c.val.(int64)
	v += val
	return c.set(v)
}

// MustInc is Inc that panics on failure.
func (c *PCPCounter) MustInc(val int64) {
	if err := c.Inc(val); err != nil {
		panic(err)
	}
}

// Up increases the counter by 1.
func (c *PCPCounter) Up() { c.MustInc(1) }

///////////////////////////////////////////////////////////////////////////////

// Gauge defines a metric that holds a single double value that can be
// incremented or decremented.
type Gauge interface {
	Metric

	Val() float64

	Set(float64) error
	MustSet(float64)

	Inc(float64) error
	Dec(float64) error

	MustInc(float64)
	MustDec(float64)
}

///////////////////////////////////////////////////////////////////////////////

// PCPGauge defines a PCP compatible Gauge metric
type PCPGauge struct {
	*pcpSingletonMetric
	mutex sync.RWMutex
}

// NewPCPGauge creates a new PCPGauge instance.
// Tt requires an initial float64 value and a metric name for construction.
// Optionally it can also take a couple of description strings that are used as
// short and long descriptions respectively.
// Internally it creates a PCP SingletonMetric with DoubleType, InstantSemantics
// and CountUnit.
func NewPCPGauge(val float64, name string, desc ...string) (*PCPGauge, error) {
	d, err := newpcpMetricDesc(name, DoubleType, InstantSemantics, OneUnit, desc...)
	if err != nil {
		return nil, err
	}

	sm, err := newpcpSingletonMetric(val, d)
	if err != nil {
		return nil, err
	}

	return &PCPGauge{sm, sync.RWMutex{}}, nil
}

// Val returns the current value of the Gauge.
func (g *PCPGauge) Val() float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.val.(float64)
}

// Set sets the current value of the Gauge.
func (g *PCPGauge) Set(val float64) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.set(val)
}

// MustSet will panic if Set fails.
func (g *PCPGauge) MustSet(val float64) {
	if err := g.Set(val); err != nil {
		panic(err)
	}
}

// Inc adds a value to the existing Gauge value.
func (g *PCPGauge) Inc(val float64) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	if val == 0 {
		return nil
	}

	v := g.val.(float64)
	return g.set(v + val)
}

// MustInc will panic if Inc fails.
func (g *PCPGauge) MustInc(val float64) {
	if err := g.Inc(val); err != nil {
		panic(err)
	}
}

// Dec adds a value to the existing Gauge value.
func (g *PCPGauge) Dec(val float64) error {
	return g.Inc(-val)
}

// MustDec will panic if Dec fails.
func (g *PCPGauge) MustDec(val float64) {
	if err := g.Dec(val); err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////

// Timer defines a metric that accumulates time periods
// Start signals the beginning of monitoring.
// End signals the end of monitoring and adding the elapsed time to the
// accumulated time, and returning it.
type Timer interface {
	Metric

	Start() error
	Stop() (float64, error)
}

///////////////////////////////////////////////////////////////////////////////

// PCPTimer implements a PCP compatible Timer
// It also functionally implements a metric with elapsed type from PCP
type PCPTimer struct {
	*pcpSingletonMetric
	mutex   sync.Mutex
	started bool
	since   time.Time
}

// NewPCPTimer creates a new PCPTimer instance of the specified unit.
// It requires a metric name and a TimeUnit for construction.
// It can optionally take a couple of description strings.
// Internally it uses a PCP SingletonMetric.
func NewPCPTimer(name string, unit TimeUnit, desc ...string) (*PCPTimer, error) {
	d, err := newpcpMetricDesc(name, DoubleType, DiscreteSemantics, unit, desc...)
	if err != nil {
		return nil, err
	}

	sm, err := newpcpSingletonMetric(float64(0), d)
	if err != nil {
		return nil, err
	}

	return &PCPTimer{sm, sync.Mutex{}, false, time.Time{}}, nil
}

// Reset resets the timer to 0
func (t *PCPTimer) Reset() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.started {
		return errors.New("trying to reset an already started timer")
	}

	return t.set(float64(0))
}

// Start signals the timer to start monitoring.
func (t *PCPTimer) Start() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.started {
		return errors.New("trying to start an already started timer")
	}

	t.since = time.Now()
	t.started = true
	return nil
}

// Stop signals the timer to end monitoring and return elapsed time so far.
func (t *PCPTimer) Stop() (float64, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.started {
		return 0, errors.New("trying to stop a stopped timer")
	}

	d := time.Since(t.since)

	var inc float64
	switch t.pcpMetricDesc.Unit() {
	case NanosecondUnit:
		inc = float64(d.Nanoseconds())
	case MicrosecondUnit:
		inc = float64(d.Nanoseconds()) * 1e-3
	case MillisecondUnit:
		inc = float64(d.Nanoseconds()) * 1e-6
	case SecondUnit:
		inc = d.Seconds()
	case MinuteUnit:
		inc = d.Minutes()
	case HourUnit:
		inc = d.Hours()
	}

	v := t.val.(float64)

	err := t.set(v + inc)
	if err != nil {
		return -1, err
	}

	t.started = false
	return v + inc, nil
}

///////////////////////////////////////////////////////////////////////////////

type instanceValue struct {
	val    interface{}
	update updateClosure
}

func newinstanceValue(val interface{}) *instanceValue {
	return &instanceValue{val, nil}
}

// pcpInstanceMetric represents a PCPMetric that can have multiple values
// over multiple instances in an instance domain.
type pcpInstanceMetric struct {
	*pcpMetricDesc
	indom *PCPInstanceDomain
	vals  map[string]*instanceValue
}

// newpcpInstanceMetric creates a new instance of PCPSingletonMetric.
func newpcpInstanceMetric(vals Instances, indom *PCPInstanceDomain, desc *pcpMetricDesc) (*pcpInstanceMetric, error) {
	if len(vals) != indom.InstanceCount() {
		return nil, errors.New("values for all instances in the instance domain only should be passed")
	}

	mvals := make(map[string]*instanceValue)

	for name := range indom.instances {
		val, present := vals[name]
		if !present {
			return nil, errors.Errorf("Instance %v not initialized", name)
		}

		if !desc.t.IsCompatible(val) {
			return nil, errors.Errorf("value %v is incompatible with type %v for Instance %v", val, desc.t, name)
		}

		val = desc.t.resolve(val)
		mvals[name] = newinstanceValue(val)
	}

	return &pcpInstanceMetric{desc, indom, mvals}, nil
}

func (m *pcpInstanceMetric) valInstance(instance string) (interface{}, error) {
	if !m.indom.HasInstance(instance) {
		return nil, errors.Errorf("%v is not an instance of this metric", instance)
	}

	return m.vals[instance].val, nil
}

// setInstance sets the value for a particular instance of the metric.
func (m *pcpInstanceMetric) setInstance(val interface{}, instance string) error {
	if !m.t.IsCompatible(val) {
		return errors.New("the value is incompatible with this metrics MetricType")
	}

	if !m.indom.HasInstance(instance) {
		return errors.Errorf("%v is not an instance of this metric", instance)
	}

	val = m.t.resolve(val)

	if m.vals[instance].val != val {
		if m.vals[instance].update != nil {
			err := m.vals[instance].update(val)
			if err != nil {
				return err
			}
		}

		m.vals[instance].val = val
	}

	return nil
}

// Indom returns the instance domain for the metric.
func (m *pcpInstanceMetric) Indom() *PCPInstanceDomain { return m.indom }

// Instances returns a slice containing all instances in the InstanceMetric.
// Basically a shorthand for metric.Indom().Instances().
func (m *pcpInstanceMetric) Instances() []string { return m.indom.Instances() }

///////////////////////////////////////////////////////////////////////////////

// PCPInstanceMetric represents a PCPMetric that can have multiple values
// over multiple instances in an instance domain.
type PCPInstanceMetric struct {
	*pcpInstanceMetric
	mutex sync.RWMutex
}

// NewPCPInstanceMetric creates a new instance of PCPSingletonMetric.
// it takes 2 extra optional strings as short and long description parameters,
// which on not being present are set to empty strings.
func NewPCPInstanceMetric(vals Instances, name string, indom *PCPInstanceDomain, t MetricType, s MetricSemantics, u MetricUnit, desc ...string) (*PCPInstanceMetric, error) {
	d, err := newpcpMetricDesc(name, t, s, u, desc...)
	if err != nil {
		return nil, err
	}

	im, err := newpcpInstanceMetric(vals, indom, d)
	if err != nil {
		return nil, err
	}

	return &PCPInstanceMetric{im, sync.RWMutex{}}, nil
}

// ValInstance returns the value for a particular instance of the metric.
func (m *PCPInstanceMetric) ValInstance(instance string) (interface{}, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.valInstance(instance)
}

// SetInstance sets the value for a particular instance of the metric.
func (m *PCPInstanceMetric) SetInstance(val interface{}, instance string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.setInstance(val, instance)
}

// MustSetInstance is a SetInstance that panics.
func (m *PCPInstanceMetric) MustSetInstance(val interface{}, instance string) {
	if err := m.SetInstance(val, instance); err != nil {
		panic(err)
	}
}

///////////////////////////////////////////////////////////////////////////////

// CounterVector defines a Counter on multiple instances.
type CounterVector interface {
	Metric

	Val(string) (int64, error)

	Set(int64, string) error
	MustSet(int64, string)
	SetAll(int64)

	Inc(int64, string) error
	MustInc(int64, string)
	IncAll(int64)

	Up(string)
	UpAll()
}

///////////////////////////////////////////////////////////////////////////////

func generateInstanceMetric(vals map[string]interface{}, name string, instances []string, t MetricType, s MetricSemantics, u MetricUnit, desc ...string) (*pcpInstanceMetric, error) {
	indomname := name + ".indom"
	indom, err := NewPCPInstanceDomain(indomname, instances)
	if err != nil {
		return nil, errors.Errorf("cannot create indom, error: %v", err)
	}

	d, err := newpcpMetricDesc(name, t, s, u, desc...)
	if err != nil {
		return nil, err
	}

	return newpcpInstanceMetric(vals, indom, d)
}

// PCPCounterVector implements a CounterVector
type PCPCounterVector struct {
	*pcpInstanceMetric
	mutex sync.RWMutex
}

// NewPCPCounterVector creates a new instance of a PCPCounterVector.
// it requires a metric name and a set of instance names and values as a map.
// it can optionally accept a couple of strings as short and long descriptions
// of the metric.
// Internally it uses a PCP InstanceMetric with Int64Type, CounterSemantics and CountUnit.
func NewPCPCounterVector(values map[string]int64, name string, desc ...string) (*PCPCounterVector, error) {
	vals := make(Instances)
	for k, v := range values {
		vals[k] = v
	}

	im, err := generateInstanceMetric(vals, name, vals.Keys(), Int64Type, CounterSemantics, OneUnit, desc...)
	if err != nil {
		return nil, err
	}

	return &PCPCounterVector{im, sync.RWMutex{}}, nil
}

// Val returns the value of a particular instance of PCPCounterVector.
func (c *PCPCounterVector) Val(instance string) (int64, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	v, err := c.valInstance(instance)
	if err != nil {
		return 0, err
	}

	return v.(int64), nil
}

// Set sets the value of a particular instance of PCPCounterVector.
func (c *PCPCounterVector) Set(val int64, instance string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	v, err := c.valInstance(instance)
	if err != nil {
		return err
	}

	if val < v.(int64) {
		return errors.Errorf("cannot set instance %s to a lesser value %v", instance, val)
	}

	return c.setInstance(val, instance)
}

// MustSet panics if Set fails.
func (c *PCPCounterVector) MustSet(val int64, instance string) {
	if err := c.Set(val, instance); err != nil {
		panic(err)
	}
}

// SetAll sets all instances to the same value and panics on an error.
func (c *PCPCounterVector) SetAll(val int64) {
	for ins := range c.indom.instances {
		c.MustSet(val, ins)
	}
}

// Inc increments the value of a particular instance of PCPCounterVector.
func (c *PCPCounterVector) Inc(inc int64, instance string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if inc < 0 {
		return errors.New("increment cannot be negative")
	}

	if inc == 0 {
		return nil
	}

	v, err := c.valInstance(instance)
	if err != nil {
		return err
	}

	return c.setInstance(v.(int64)+inc, instance)
}

// MustInc panics if Inc fails.
func (c *PCPCounterVector) MustInc(inc int64, instance string) {
	if err := c.Inc(inc, instance); err != nil {
		panic(err)
	}
}

// IncAll increments all instances by the same value and panics on an error.
func (c *PCPCounterVector) IncAll(val int64) {
	for ins := range c.indom.instances {
		c.MustInc(val, ins)
	}
}

// Up increments the value of a particular instance ny 1.
func (c *PCPCounterVector) Up(instance string) { c.MustInc(1, instance) }

// UpAll ups all instances and panics on an error.
func (c *PCPCounterVector) UpAll() { c.IncAll(1) }

///////////////////////////////////////////////////////////////////////////////

// GaugeVector defines a Gauge on multiple instances
type GaugeVector interface {
	Metric

	Val(string) (float64, error)

	Set(float64, string) error
	MustSet(float64, string)
	SetAll(float64)

	Inc(float64, string) error
	MustInc(float64, string)
	IncAll(float64)

	Dec(float64, string) error
	MustDec(float64, string)
	DecAll(float64)
}

///////////////////////////////////////////////////////////////////////////////

// PCPGaugeVector implements a GaugeVector
type PCPGaugeVector struct {
	*pcpInstanceMetric
	mutex sync.RWMutex
}

// NewPCPGaugeVector creates a new instance of a PCPGaugeVector.
// It requires a name and map of instance names to their values.
// Optionally, it can also accept a couple of strings providing more details
// about the metric.
func NewPCPGaugeVector(values map[string]float64, name string, desc ...string) (*PCPGaugeVector, error) {
	vals := make(Instances)
	for k, v := range values {
		vals[k] = v
	}

	im, err := generateInstanceMetric(vals, name, vals.Keys(), DoubleType, InstantSemantics, OneUnit, desc...)
	if err != nil {
		return nil, err
	}

	return &PCPGaugeVector{im, sync.RWMutex{}}, nil
}

// Val returns the value of a particular instance of PCPGaugeVector
func (g *PCPGaugeVector) Val(instance string) (float64, error) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	val, err := g.valInstance(instance)
	if err != nil {
		return 0, err
	}

	return val.(float64), nil
}

// Set sets the value of a particular instance of PCPGaugeVector
func (g *PCPGaugeVector) Set(val float64, instance string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.setInstance(val, instance)
}

// MustSet panics if Set fails
func (g *PCPGaugeVector) MustSet(val float64, instance string) {
	if err := g.Set(val, instance); err != nil {
		panic(err)
	}
}

// SetAll sets all instances to the same value and panics on an error
func (g *PCPGaugeVector) SetAll(val float64) {
	for ins := range g.indom.instances {
		g.MustSet(val, ins)
	}
}

// Inc increments the value of a particular instance of PCPGaugeVector
func (g *PCPGaugeVector) Inc(inc float64, instance string) error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	v, err := g.valInstance(instance)
	if err != nil {
		return err
	}

	return g.setInstance(v.(float64)+inc, instance)
}

// MustInc panics if Inc fails
func (g *PCPGaugeVector) MustInc(inc float64, instance string) {
	if err := g.Inc(inc, instance); err != nil {
		panic(err)
	}
}

// IncAll increments all instances by the same value and panics on an error
func (g *PCPGaugeVector) IncAll(val float64) {
	for ins := range g.indom.instances {
		g.MustInc(val, ins)
	}
}

// Dec decrements the value of a particular instance of PCPGaugeVector
func (g *PCPGaugeVector) Dec(inc float64, instance string) error { return g.Inc(-inc, instance) }

// MustDec panics if Dec fails
func (g *PCPGaugeVector) MustDec(inc float64, instance string) { g.MustInc(-inc, instance) }

// DecAll decrements all instances by the same value and panics on an error
func (g *PCPGaugeVector) DecAll(val float64) { g.IncAll(-val) }

///////////////////////////////////////////////////////////////////////////////

// Histogram defines a metric that records a distribution of data
type Histogram interface {
	Max() int64 // Maximum value recorded so far
	Min() int64 // Minimum value recorded so far

	High() int64 // Highest allowed value
	Low() int64  // Lowest allowed value

	Record(int64) error         // Records a new value
	RecordN(int64, int64) error // Records multiple instances of the same value

	MustRecord(int64)
	MustRecordN(int64, int64)

	Mean() float64              // Mean of all recorded data
	Variance() float64          // Variance of all recorded data
	StandardDeviation() float64 // StandardDeviation of all recorded data
	Percentile(float64) int64   // Percentile returns the value at the passed percentile
}

///////////////////////////////////////////////////////////////////////////////

// PCPHistogram implements a histogram for PCP backed by the coda hale hdrhistogram
// https://github.com/codahale/hdrhistogram
type PCPHistogram struct {
	*pcpInstanceMetric
	mutex sync.RWMutex
	h     *histogram.Histogram
}

// the maximum and minimum values that can be recorded by a histogram
const (
	HistogramMin = 0
	HistogramMax = 3600000000
)

func normalize(low, high int64, sigfigures int) (int64, int64, int) {
	if low < HistogramMin {
		low = HistogramMin
	}

	if low > HistogramMax {
		low = HistogramMax
	}

	if high < HistogramMin {
		high = HistogramMin
	}

	if high > HistogramMax {
		high = HistogramMax
	}

	if sigfigures < 1 {
		sigfigures = 1
	}

	if sigfigures > 5 {
		sigfigures = 5
	}

	return low, high, sigfigures
}

// NewPCPHistogram returns a new instance of PCPHistogram.
// The lowest value for `low` is 0.
// The highest value for `high` is 3,600,000,000.
// `low` **must** be less than `high`.
// The value of `sigfigures` can be between 1 and 5.
// It also requires a unit to be explicitly passed for construction.
// Optionally, a couple of description strings may be passed as the short and
// long descriptions of the metric.
func NewPCPHistogram(name string, low, high int64, sigfigures int, unit MetricUnit, desc ...string) (*PCPHistogram, error) {
	if low > high {
		return nil, errors.New("low cannot be larger than high")
	}

	low, high, sigfigures = normalize(low, high, sigfigures)

	h := histogram.New(low, high, sigfigures)

	vals := make(Instances)
	for _, s := range histogramInstances {
		vals[s] = float64(0)
	}

	d, err := newpcpMetricDesc(name, DoubleType, InstantSemantics, unit, desc...)
	if err != nil {
		return nil, err
	}

	m, err := newpcpInstanceMetric(vals, histogramIndom, d)
	if err != nil {
		return nil, err
	}

	return &PCPHistogram{m, sync.RWMutex{}, h}, nil
}

// High returns the maximum recordable value.
func (h *PCPHistogram) High() int64 { return h.h.LowestTrackableValue() }

// Low returns the minimum recordable value.
func (h *PCPHistogram) Low() int64 { return h.h.HighestTrackableValue() }

// Max returns the maximum recorded value so far.
func (h *PCPHistogram) Max() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return int64(h.vals["max"].val.(float64))
}

// Min returns the minimum recorded value so far.
func (h *PCPHistogram) Min() int64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return int64(h.vals["min"].val.(float64))
}

func (h *PCPHistogram) update() error {
	updateinstance := func(instance string, val float64) error {
		if h.vals[instance].val != val {
			return h.setInstance(val, instance)
		}
		return nil
	}

	if err := updateinstance("min", float64(h.h.Min())); err != nil {
		return err
	}

	if err := updateinstance("max", float64(h.h.Max())); err != nil {
		return err
	}

	if err := updateinstance("mean", h.h.Mean()); err != nil {
		return err
	}

	stddev := h.h.StdDev()

	if err := updateinstance("standard_deviation", stddev); err != nil {
		return err
	}

	if err := updateinstance("variance", stddev*stddev); err != nil {
		return err
	}

	return nil
}

// Record records a new value.
func (h *PCPHistogram) Record(val int64) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	err := h.h.RecordValue(val)
	if err != nil {
		return err
	}

	return h.update()
}

// MustRecord panics if Record fails.
func (h *PCPHistogram) MustRecord(val int64) {
	if err := h.Record(val); err != nil {
		panic(err)
	}
}

// RecordN records multiple instances of the same value.
func (h *PCPHistogram) RecordN(val, n int64) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	err := h.h.RecordValues(val, n)
	if err != nil {
		return err
	}

	return h.update()
}

// MustRecordN panics if RecordN fails.
func (h *PCPHistogram) MustRecordN(val, n int64) {
	if err := h.RecordN(val, n); err != nil {
		panic(err)
	}
}

// Mean returns the mean of all values recorded so far.
func (h *PCPHistogram) Mean() float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.vals["mean"].val.(float64)
}

// StandardDeviation returns the standard deviation of all values recorded so far.
func (h *PCPHistogram) StandardDeviation() float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.vals["standard_deviation"].val.(float64)
}

// Variance returns the variance of all values recorded so far.
func (h *PCPHistogram) Variance() float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.vals["variance"].val.(float64)
}

// Percentile returns the value at the passed percentile.
func (h *PCPHistogram) Percentile(p float64) int64 { return h.h.ValueAtQuantile(p) }

// HistogramBucket is a single histogram bucket within a fixed range.
type HistogramBucket struct {
	From, To, Count int64
}

// Buckets returns a list of histogram buckets.
func (h *PCPHistogram) Buckets() []*HistogramBucket {
	b := h.h.Distribution()
	buckets := make([]*HistogramBucket, len(b))
	for i := 0; i < len(b); i++ {
		buckets[i] = &HistogramBucket{b[i].From, b[i].To, b[i].Count}
	}
	return buckets
}
