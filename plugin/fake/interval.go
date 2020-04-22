package fake

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include "plugin.h"
//
// static cdtime_t interval = TIME_T_TO_CDTIME_T_STATIC(10);
// cdtime_t plugin_get_interval(void) {
//   return interval;
// }
// static void plugin_set_interval(cdtime_t d) {
//   interval = d;
// }
import "C"

import (
	"time"

	"collectd.org/cdtime"
)

// SetInterval sets the interval returned by the fake plugin_get_interval()
// function.
func SetInterval(d time.Duration) {
	ival := cdtime.NewDuration(d)
	C.plugin_set_interval(C.cdtime_t(ival))
}
