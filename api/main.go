// Package api defines data types representing core collectd data types.
package api // import "collectd.org/api"

import (
	"time"
)

// Value represents either a Gauge or a Derive. It is Go's equivalent to the C
// union value_t. If a function accepts a Value, you may pass in either a Gauge
// or a Derive. Passing in any other type may or may not panic.
type Value interface {
	Type() string
}

// Gauge represents a gauge metric value, such as a temperature.
// This is Go's equivalent to the C type "gauge_t".
type Gauge float64

// Type returns "gauge".
func (v Gauge) Type() string { return "gauge" }

// Derive represents a counter metric value, such as bytes sent over the
// network. When the counter wraps around (overflows) or is reset, this is
// interpreted as a (huge) negative rate, which is discarded.
// This is Go's equivalent to the C type "derive_t".
type Derive int64

// Type returns "derive".
func (v Derive) Type() string { return "derive" }

// Counter represents a counter metric value, such as bytes sent over the
// network. When a counter value is smaller than the previous value, a wrap
// around (overflow) is assumed. This causes huge spikes in case a counter is
// reset. Only use Counter for very specific cases. If in doubt, use Derive
// instead.
// This is Go's equivalent to the C type "counter_t".
type Counter uint64

// Type returns "counter".
func (v Counter) Type() string { return "counter" }

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

// Writer are objects accepting a ValueList for writing, for example to the
// network.
type Writer interface {
	Write(vl ValueList) error
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
