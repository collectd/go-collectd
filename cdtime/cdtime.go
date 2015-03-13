/*
Package cdtime implements methods to convert from and to collectd's internal time
representation, cdtime_t.
*/
package cdtime // import "collectd.org/cdtime"

import (
	"time"
)

// Time represens a time in collectd's internal representation.
type Time uint64

// New returns a new Time representing time t.
func New(t time.Time) Time {
	return newNano(uint64(t.UnixNano()))
}

// NewDuration returns a new Time representing duration d.
func NewDuration(d time.Duration) Time {
	return newNano(uint64(d.Nanoseconds()))
}

// Time converts and returns the time as time.Time.
func (t Time) Time() time.Time {
	s := int64(t >> 30)
	ns := (int64(t&0x3fffffff) * 1000000000) >> 30

	return time.Unix(s, ns)
}

// Duration converts and returns the duration as time.Time.
func (t Time) Duration() time.Duration {
	s := int64(t >> 30)
	ns := (int64(t&0x3fffffff) * 1000000000) >> 30

	return time.Duration(1000000000*s+ns) * time.Nanosecond
}

func newNano(ns uint64) Time {
	// break into seconds and nano-seconds so the left-shift doesn't overflow.
	s := ns / 1000000000
	ns = ns % 1000000000

	return Time((s << 30) | ((ns << 30) / 1000000000))
}
