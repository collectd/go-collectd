// +build go1.5,cgo

/*
Package plugin exports the functions required to write collectd plugins in Go.

This package provides the abstraction necessary to write plugins for collectd
in Go, compile them into a shared object and let the daemon load and use them.

Example plugin

To understand how this module is being used, please consider the following
example:

  package main

  import (
          "context"
          "fmt"
          "time"

          "collectd.org/api"
          "collectd.org/plugin"
  )

  type examplePlugin struct{}

  func (examplePlugin) Read(ctx context.Context) error {
          vl := &api.ValueList{
                  Identifier: api.Identifier{
                          Host:   "example.com",
                          Plugin: "goplug",
                          Type:   "gauge",
                  },
                  Time:     time.Now(),
                  Interval: 10 * time.Second,
                  Values:   []api.Value{api.Gauge(42)},
                  DSNames:  []string{"value"},
          }
          if err := plugin.Write(ctx, vl); err != nil {
                  return fmt.Errorf("plugin.Write: %w", err)
          }

          return nil
  }

  func init() {
          plugin.RegisterRead("example", examplePlugin{})
  }

  func main() {} // ignored

The first step when writing a new plugin with this package, is to create a new
"main" package. Even though it has to have a main() function to make cgo happy,
the main() function is ignored. Instead, put your startup code into the init()
function which essentially takes on the same role as the module_register()
function in C based plugins.

Then, define a type which implements the Reader interface by implementing the
"Read() error" function. In the example above, this type is called
"examplePlugin". Create an instance of this type and pass it to RegisterRead()
in the init() function.

Build flags

To compile your plugin, set up the CGO_CPPFLAGS environment variable and call
"go build" with the following options:

  export COLLECTD_SRC="/path/to/collectd"
  export CGO_CPPFLAGS="-I${COLLECTD_SRC}/src/daemon -I${COLLECTD_SRC}/src"
  go build -buildmode=c-shared -o example.so
*/
package plugin // import "collectd.org/plugin"

// #cgo CPPFLAGS: -DHAVE_CONFIG_H
// #cgo LDFLAGS: -ldl
// #include <stdlib.h>
// #include <dlfcn.h>
// #include "plugin.h"
//
// int dispatch_values_wrapper (value_list_t const *vl);
// cdtime_t plugin_get_interval_wrapper(void);
//
// data_source_t *ds_dsrc(data_set_t const *ds, size_t i);
//
// void value_list_add_counter (value_list_t *, counter_t);
// void value_list_add_derive  (value_list_t *, derive_t);
// void value_list_add_gauge   (value_list_t *, gauge_t);
// counter_t value_list_get_counter (value_list_t *, size_t);
// derive_t  value_list_get_derive  (value_list_t *, size_t);
// gauge_t   value_list_get_gauge   (value_list_t *, size_t);
//
// int register_read_wrapper (char const *group, char const *name,
//     plugin_read_cb callback,
//     cdtime_t interval,
//     user_data_t *ud);
// int wrap_read_callback(user_data_t *);
//
// int register_write_wrapper (char const *, plugin_write_cb, user_data_t *);
// int wrap_write_callback(data_set_t *, value_list_t *, user_data_t *);
//
// int register_shutdown_wrapper (char *, plugin_shutdown_cb);
// int wrap_shutdown_callback(void);
//
// int register_log_wrapper(char const *, plugin_log_cb, user_data_t const *);
// int wrap_log_callback(int, char *, user_data_t *);
//
// typedef int (*plugin_complex_config_cb)(oconfig_item_t *);
//
// int register_complex_config_wrapper(char const *, plugin_complex_config_cb);
// int wrap_configure_callback(oconfig_item_t *);
//
// int register_init_wrapper (const char *name, plugin_init_cb callback);
//
// typedef void (*free_func_t)(void *);
import "C"

import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"collectd.org/api"
	"collectd.org/cdtime"
)

// Reader defines the interface for read callbacks, i.e. Go functions that are
// called periodically from the collectd daemon.
type Reader interface {
	Read(ctx context.Context) error
}

func strcpy(dst []C.char, src string) {
	byteStr := []byte(src)
	cStr := make([]C.char, len(byteStr)+1)

	for i, b := range byteStr {
		cStr[i] = C.char(b)
	}
	cStr[len(cStr)-1] = C.char(0)

	copy(dst, cStr)
}

func newValueListT(vl *api.ValueList) (*C.value_list_t, error) {
	ret := &C.value_list_t{}

	strcpy(ret.host[:], vl.Host)
	strcpy(ret.plugin[:], vl.Plugin)
	strcpy(ret.plugin_instance[:], vl.PluginInstance)
	strcpy(ret._type[:], vl.Type)
	strcpy(ret.type_instance[:], vl.TypeInstance)
	ret.interval = C.cdtime_t(cdtime.NewDuration(vl.Interval))
	ret.time = C.cdtime_t(cdtime.New(vl.Time))

	for _, v := range vl.Values {
		switch v := v.(type) {
		case api.Counter:
			if _, err := C.value_list_add_counter(ret, C.counter_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_counter: %w", err)
			}
		case api.Derive:
			if _, err := C.value_list_add_derive(ret, C.derive_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_derive: %w", err)
			}
		case api.Gauge:
			if _, err := C.value_list_add_gauge(ret, C.gauge_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_gauge: %w", err)
			}
		default:
			return nil, fmt.Errorf("not yet supported: %T", v)
		}
	}

	return ret, nil
}

// Writer implements the api.Write interface.
type Writer struct{}

// Write implements the api.Writer interface for the collectd daemon.
func (Writer) Write(ctx context.Context, vl *api.ValueList) error {
	return Write(ctx, vl)
}

// Write converts a ValueList and calls the plugin_dispatch_values() function
// of the collectd daemon.
func Write(ctx context.Context, vl *api.ValueList) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	vlt, err := newValueListT(vl)
	if err != nil {
		return err
	}
	defer C.free(unsafe.Pointer(vlt.values))

	status, err := C.dispatch_values_wrapper(vlt)
	return wrapCError(status, err, "dispatch_values")
}

// readFuncs holds references to all read callbacks, so the garbage collector
// doesn't get any funny ideas.
var readFuncs = make(map[string]Reader)

// RegisterRead registers a new read function with the daemon which is called
// periodically.
func RegisterRead(name string, r Reader) error {
	cGroup := C.CString("golang")
	defer C.free(unsafe.Pointer(cGroup))

	cName := C.CString(name)
	ud := C.user_data_t{
		data:      unsafe.Pointer(cName),
		free_func: C.free_func_t(C.free),
	}

	status, err := C.register_read_wrapper(cGroup, cName,
		C.plugin_read_cb(C.wrap_read_callback),
		C.cdtime_t(0),
		&ud)
	if err := wrapCError(status, err, "register_read"); err != nil {
		return err
	}

	readFuncs[name] = r
	return nil
}

type key struct{}

var nameKey key

func withName(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, nameKey, name)
}

// Name returns the name of the plugin / callback.
func Name(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(nameKey).(string)
	return name, ok
}

//export wrap_read_callback
func wrap_read_callback(ud *C.user_data_t) C.int {
	name := C.GoString((*C.char)(ud.data))
	r, ok := readFuncs[name]
	if !ok {
		return -1
	}

	ctx := withName(context.Background(), name)
	if err := r.Read(ctx); err != nil {
		Errorf("%s plugin: Read() failed: %v", name, err)
		return -1
	}

	return 0
}

// Interval returns the interval in which read callbacks are being called. May
// only be called from within a read callback.
func Interval() (time.Duration, error) {
	ival, err := C.plugin_get_interval_wrapper()
	if err != nil {
		return 0, fmt.Errorf("plugin_get_interval() failed: %w", err)
	}

	return cdtime.Time(ival).Duration(), nil
}

// writeFuncs holds references to all write callbacks, so the garbage collector
// doesn't get any funny ideas.
var writeFuncs = make(map[string]api.Writer)

// RegisterWrite registers a new write function with the daemon which is called
// for every metric collected by collectd.
//
// Please note that multiple threads may call this function concurrently. If
// you're accessing shared resources, such as a memory buffer, you have to
// implement appropriate locking around these accesses.
func RegisterWrite(name string, w api.Writer) error {
	cName := C.CString(name)
	ud := C.user_data_t{
		data:      unsafe.Pointer(cName),
		free_func: C.free_func_t(C.free),
	}

	status, err := C.register_write_wrapper(cName, C.plugin_write_cb(C.wrap_write_callback), &ud)
	if err := wrapCError(status, err, "register_write"); err != nil {
		return err
	}

	writeFuncs[name] = w
	return nil
}

//export wrap_write_callback
func wrap_write_callback(ds *C.data_set_t, cvl *C.value_list_t, ud *C.user_data_t) C.int {
	name := C.GoString((*C.char)(ud.data))
	w, ok := writeFuncs[name]
	if !ok {
		return -1
	}

	vl := &api.ValueList{
		Identifier: api.Identifier{
			Host:           C.GoString(&cvl.host[0]),
			Plugin:         C.GoString(&cvl.plugin[0]),
			PluginInstance: C.GoString(&cvl.plugin_instance[0]),
			Type:           C.GoString(&cvl._type[0]),
			TypeInstance:   C.GoString(&cvl.type_instance[0]),
		},
		Time:     cdtime.Time(cvl.time).Time(),
		Interval: cdtime.Time(cvl.interval).Duration(),
	}

	// TODO: Remove 'size_t' cast on 'ds_num' upon 5.7 release.
	for i := C.size_t(0); i < C.size_t(ds.ds_num); i++ {
		dsrc := C.ds_dsrc(ds, i)

		switch dsrc._type {
		case C.DS_TYPE_COUNTER:
			v := C.value_list_get_counter(cvl, i)
			vl.Values = append(vl.Values, api.Counter(v))
		case C.DS_TYPE_DERIVE:
			v := C.value_list_get_derive(cvl, i)
			vl.Values = append(vl.Values, api.Derive(v))
		case C.DS_TYPE_GAUGE:
			v := C.value_list_get_gauge(cvl, i)
			vl.Values = append(vl.Values, api.Gauge(v))
		default:
			Errorf("%s plugin: data source type %d is not supported", name, dsrc._type)
			return -1
		}

		vl.DSNames = append(vl.DSNames, C.GoString(&dsrc.name[0]))
	}

	ctx := withName(context.Background(), name)
	if err := w.Write(ctx, vl); err != nil {
		Errorf("%s plugin: Write() failed: %v", name, err)
		return -1
	}

	return 0
}

// First declare some types, interfaces, general functions

// Shutter is called to shut down the plugin gracefully.
type Shutter interface {
	Shutdown(context.Context) error
}

// shutdownFuncs holds references to all shutdown callbacks
var shutdownFuncs = make(map[string]Shutter)

//export wrap_shutdown_callback
func wrap_shutdown_callback() C.int {
	ret := C.int(0)
	for name, f := range shutdownFuncs {
		ctx := withName(context.Background(), name)
		if err := f.Shutdown(ctx); err != nil {
			Errorf("%s plugin: Shutdown() failed: %v", name, err)
			ret = -1
		}
	}
	return ret
}

// RegisterShutdown registers a shutdown function with the daemon which is called
// when the plugin is required to shutdown gracefully.
func RegisterShutdown(name string, s Shutter) error {
	// Only register a callback the first time one is implemented, subsequent
	// callbacks get added to a map and called sequentially from the same
	// (C) callback.
	if len(shutdownFuncs) <= 0 {
		cName := C.CString(name)
		defer C.free(unsafe.Pointer(cName))

		status, err := C.register_shutdown_wrapper(cName, C.plugin_shutdown_cb(C.wrap_shutdown_callback))
		if err := wrapCError(status, err, "register_shutdown"); err != nil {
			return err
		}
	}
	shutdownFuncs[name] = s
	return nil
}

// Logger implements a logging callback.
type Logger interface {
	Log(context.Context, Severity, string)
}

// RegisterLog registers a logging function with the daemon which is called
// whenever a log message is generated.
func RegisterLog(name string, l Logger) error {
	cName := C.CString(name)
	ud := C.user_data_t{
		data:      unsafe.Pointer(cName),
		free_func: C.free_func_t(C.free),
	}

	status, err := C.register_log_wrapper(cName, C.plugin_log_cb(C.wrap_log_callback), &ud)
	if err := wrapCError(status, err, "register_log"); err != nil {
		return err
	}

	logFuncs[name] = l
	return nil
}

var logFuncs = make(map[string]Logger)

//export wrap_log_callback
func wrap_log_callback(sev C.int, msg *C.char, ud *C.user_data_t) C.int {
	name := C.GoString((*C.char)(ud.data))
	f, ok := logFuncs[name]
	if !ok {
		return -1
	}

	ctx := withName(context.Background(), name)
	f.Log(ctx, Severity(sev), C.GoString(msg))

	return 0
}

// Configurer implements a Configure callback.
type Configurer interface {
	Configure(context.Context, interface{})
}

// Configurers are registered once but Configs may be received multiple times and merged together before unmarshalling,
// so they're tracked together for a convenient Unmarshal call.
type configPair struct {
	f Configurer
	c Config
}

var configureFuncs = make(map[string]configPair)

// RegisterConfigure registers a configuration-receiving function with the daemon.
func RegisterConfigure(name string, c Configurer) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	status, err := C.register_complex_config_wrapper(cName, C.plugin_complex_config_cb(C.wrap_configure_callback))
	if err := wrapCError(status, err, "register_configure"); err != nil {
		return err
	}

	configureFuncs[name] = configPair{
		f: c,
	}
	return nil
}

//export wrap_configure_callback
func wrap_configure_callback(ci *C.oconfig_item_t) C.int {
	panic("Not yet implemented")
	return 0
}

func wrapCError(status C.int, err error, name string) error {
	if err != nil {
		return fmt.Errorf("%s failed: %w", name, err)
	}
	if status != 0 {
		return fmt.Errorf("%s failed with status %d", name, status)
	}
	return nil
}

//export module_register
func module_register() {
}
