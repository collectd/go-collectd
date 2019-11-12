// Package api defines data types representing core collectd data types.
package api

import (
	"C"
	"time"
)

/*
 * "cdtime_t" is a 64bit unsigned integer. The time is stored at a 2^-30 second
 * resolution, i.e. the most significant 34 bit are used to store the time in
 * seconds, the least significant bits store the sub-second part in something
 * very close to nanoseconds. *The* big advantage of storing time in this
 * manner is that comparing times and calculating differences is as simple as
 * it is with "time_t", i.e. a simple integer comparison / subtraction works.
 */
const (
	doubleToCollectdTimeTFactor = 1073741824 /* 2^30 = 1073741824 */
)

// NewCollectdTimeTAsUInt64 converts a time.Duration object
// to an `uint64` representation of collectd's C cdtime_t
func NewCollectdTimeTAsUInt64(dur time.Duration) uint64 {
	return uint64(dur.Seconds() * doubleToCollectdTimeTFactor)
}
