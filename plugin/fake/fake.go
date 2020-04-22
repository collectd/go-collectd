// Package fake implements fake versions of the C functions imported from the
// collectd daemon for testing.
package fake

// void reset_log(void);
// void reset_read(void);
// void reset_shutdown(void);
// void reset_write(void);
import "C"

import (
	"time"
)

// TearDown cleans up after a test and prepares shared resources for the next
// test.
//
// Note that this only resets the state of the fake implementations, such as
// "plugin_register_log()". The Go code in "collectd.org/plugin" may still hold
// a reference to the callback even after this function has been called.
func TearDown() {
	SetInterval(10 * time.Second)
	C.reset_log()
	C.reset_read()
	C.reset_shutdown()
	C.reset_write()
}
