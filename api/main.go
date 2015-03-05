// Package api defines data types representing core collectd types.
package api

import (
	"fmt"
	"time"
)

type Number interface {
	String() string
}

type Gauge float64

func (g Gauge) String() string {
	return fmt.Sprintf("%g", g)
}

type Derive int64

func (d Derive) String() string {
	return fmt.Sprintf("%d", d)
}

// Identifier identifies one metric.
type Identifier struct {
	Host                   string
	Plugin, PluginInstance string
	Type, TypeInstance     string
}

// ValueList represents one (set of) data point(s) of one metric.
type ValueList struct {
	Identifier
	Time     time.Time
	Interval time.Duration
	Values   []Number
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
