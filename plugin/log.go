// +build go1.5,cgo

package plugin // import "collectd.org/plugin"

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <dlfcn.h>
// #include "plugin.h"
//
// static void (*plugin_log_) (int, char const *, ...) = NULL;
// void wrap_plugin_log(int severity, char *msg) {
//   if (plugin_log_ == NULL) {
//     void *hnd = dlopen(NULL, RTLD_LAZY);
//     plugin_log_ = dlsym(hnd, "plugin_log");
//     dlclose(hnd);
//   }
//   (*plugin_log_) (severity, "%s", msg);
// }
import "C"

import (
	"fmt"
	"unsafe"
)

// Severity is the severity of log messages. These are well-known constants
// within collectd, so don't define your own. Use the constants provided by
// this package instead.
type Severity int

const (
	SeverityError   Severity = 3
	SeverityWarning Severity = 4
	SeverityNotice  Severity = 5
	SeverityInfo    Severity = 6
	SeverityDebug   Severity = 7
)

func log(s Severity, msg string) error {
	ptr := C.CString(msg)
	defer C.free(unsafe.Pointer(ptr))

	_, err := C.wrap_plugin_log(C.int(s), ptr)
	return err
}

// Error logs an error using plugin_log(). Arguments are handled in the manner
// of fmt.Print.
func Error(v ...interface{}) error {
	return log(SeverityError, fmt.Sprint(v...))
}

// Errorf logs an error using plugin_log(). Arguments are handled in the manner
// of fmt.Printf.
func Errorf(format string, v ...interface{}) error {
	return Error(fmt.Sprintf(format, v...))
}

// Warning logs a warning using plugin_log(). Arguments are handled in the
// manner of fmt.Print.
func Warning(v ...interface{}) error {
	return log(SeverityWarning, fmt.Sprint(v...))
}

// Warningf logs a warning using plugin_log(). Arguments are handled in the
// manner of fmt.Printf.
func Warningf(format string, v ...interface{}) error {
	return Warning(fmt.Sprintf(format, v...))
}

// Notice logs a notice using plugin_log(). Arguments are handled in the manner
// of fmt.Print.
func Notice(v ...interface{}) error {
	return log(SeverityNotice, fmt.Sprint(v...))
}

// Noticef logs a notice using plugin_log(). Arguments are handled in the
// manner of fmt.Printf.
func Noticef(format string, v ...interface{}) error {
	return Notice(fmt.Sprintf(format, v...))
}

// Info logs a purely informal message using plugin_log(). Arguments are
// handled in the manner of fmt.Print.
func Info(v ...interface{}) error {
	return log(SeverityInfo, fmt.Sprint(v...))
}

// Infof logs a purely informal message using plugin_log(). Arguments are
// handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) error {
	return Info(fmt.Sprintf(format, v...))
}

// Debug logs a debugging message using plugin_log(). Arguments are handled in
// the manner of fmt.Print.
func Debug(v ...interface{}) error {
	return log(SeverityDebug, fmt.Sprint(v...))
}

// Debugf logs a debugging message using plugin_log(). Arguments are handled in
// the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) error {
	return Debug(fmt.Sprintf(format, v...))
}
