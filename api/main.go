// Package api defines data types representing core collectd data types.
package api // import "collectd.org/api"

import (
	"time"
)

// Value represents either a Gauge or a Derive. It is Go's equivalent to the C
// union value_t. If a function accepts a Value, you may pass in either a Gauge
// or a Derive. Passing in any other type may or may not panic.
type Value interface{}

// Gauge represents a gauge metric value, such as a temperature. It is Go's
// equivalent to the C type gauge_t.
type Gauge float64

// Derive represents a derive or counter metric value, such as bytes sent over
// the network. It is Go's equivalent to the C type derive_t.
type Derive int64

// Identifier identifies one metric.
type Identifier struct {
	Host                   string
	Plugin, PluginInstance string
	Type, TypeInstance     string
}

// ValueList represents one (set of) data point(s) of one metric. It is Go's
// equivalent of the C type value_list_t.
type ValueList struct {
	Identifier
	Time     time.Time
	Interval time.Duration
	Values   []Value
}

// Dispatcher are objects accepting a ValueList for "dispatching", e.g. writing
// to the network.
type Dispatcher interface {
	Dispatch(vl ValueList) error
}

// String returns a string representation of the Identifier.
func (id Identifier) String() string {
	str := id.Host + "/" + id.Plugin
	if id.PluginInstance != "" {
		str += "-" + id.PluginInstance
	}
	str += "/" + id.Type
	if id.TypeInstance != "" {
		str += "-" + id.TypeInstance
	}
	return str
}

func cdtimeNano(ns uint64) uint64 {
	// break into seconds and nano-seconds so the left-shift doesn't overflow.
	s := ns / 1000000000
	ns = ns % 1000000000

	return (s << 30) | ((ns << 30) / 1000000000)
}

// Cdtime converts a time.Time to collectd's internal time format.
func Cdtime(t time.Time) uint64 {
	return cdtimeNano(uint64(t.UnixNano()))
}

// CdtimeDuration converts a time.Duration to collectd's internal time format.
func CdtimeDuration(d time.Duration) uint64 {
	return cdtimeNano(uint64(d.Nanoseconds()))
}
