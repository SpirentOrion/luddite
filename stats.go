package luddite

import "time"

type Stats interface {
	// Incr increments a counter metric. Often used to note a particular event.
	Incr(stat string, count int64) error
	// Decr decrements a counter metric. Often used to note a particular event.
	Decr(stat string, count int64) error
	// Timing tracks a duration event.
	Timing(stat string, delta int64) error
	// PrecisionTiming tracks a duration event but the time delta has to be a duration
	PrecisionTiming(stat string, delta time.Duration) error
	// Gauge represents a constant data type. They are not subject to averaging,
	// and they don’t change unless you change them. That is, once you set a gauge value,
	// it will be a flat line on the graph until you change it again.
	Gauge(stat string, value int64) error
	// GaugeDelta records a delta from the previous value (as int64).
	GaugeDelta(stat string, value int64) error
	// FGauge is a Gauge working with float64 values.
	FGauge(stat string, value float64) error
	// FGaugeDelta records a delta from the previous value (as float64).
	FGaugeDelta(stat string, value float64) error
	// Absolute - Send absolute-valued metric (not averaged/aggregated)
	Absolute(stat string, value int64) error
	// FAbsolute - Send absolute-valued metric (not averaged/aggregated)
	FAbsolute(stat string, value float64) error
	// Total - Send a metric that is continously increasing, e.g. read operations since boot
	Total(stat string, value int64) error
}

type NullStats struct{}

func (n *NullStats) Incr(stat string, count int64) error {
	return nil
}

func (n *NullStats) Decr(stat string, count int64) error {
	return nil
}

func (n *NullStats) Timing(stat string, delta int64) error {
	return nil
}

func (n *NullStats) PrecisionTiming(stat string, delta time.Duration) error {
	return nil
}

func (n *NullStats) Gauge(stat string, value int64) error {
	return nil
}

func (n *NullStats) GaugeDelta(stat string, value int64) error {
	return nil
}

func (n *NullStats) FGauge(stat string, value float64) error {
	return nil
}

func (n *NullStats) FGaugeDelta(stat string, value float64) error {
	return nil
}

func (n *NullStats) Absolute(stat string, value int64) error {
	return nil
}

func (n *NullStats) FAbsolute(stat string, value float64) error {
	return nil
}

func (n *NullStats) Total(stat string, value int64) error {
	return nil
}
