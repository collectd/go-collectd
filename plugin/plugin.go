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
	  "collectd.org/api"
	  "collectd.org/plugin"
  )

  type ExamplePlugin struct{}

  func (p *ExamplePlugin) Read() error {
	  vl := api.ValueList{
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
	  if err := plugin.Write(vl); err != nil {
		  return err
	  }

	  return nil
  }

  func init() {
	  plugin.RegisterRead("example", &ExamplePlugin{})
  }

  func main() {} // ignored

The first step when writing a new plugin with this package, is to create a new
"main" package. Even though it has to have a main() function to make cgo happy,
the main() function is ignored. Instead, put your startup code into the init()
function which essentially takes on the same role as the module_register()
function in C based plugins.

Then, define a type which implements the Reader interface by implementing the
"Read() error" function. In the example above, this type is called
ExamplePlugin. Create an instance of this type and pass it to RegisterRead() in
the init() function.

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
// int register_read_wrapper (char const *group, char const *name,
//     plugin_read_cb callback,
//     cdtime_t interval,
//     user_data_t *ud);
//
// data_source_t *ds_dsrc(data_set_t const *ds, size_t i);
//
// void value_list_add_gauge (value_list_t *vl, gauge_t g);
// void value_list_add_derive (value_list_t *vl, derive_t d);
// gauge_t value_list_get_gauge (value_list_t *, size_t);
// derive_t value_list_get_derive (value_list_t *, size_t);
//
// int wrap_read_callback(user_data_t *);
//
// int register_write_wrapper (char const *, plugin_write_cb, user_data_t *);
// int wrap_write_callback(data_set_t *, value_list_t *, user_data_t *);
import "C"

import (
	"fmt"
	"unsafe"

	"collectd.org/api"
	"collectd.org/cdtime"
)

// Reader defines the interface for read callbacks, i.e. Go functions that are
// called periodically from the collectd daemon.
type Reader interface {
	Read() error
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

func newValueListT(vl api.ValueList) (*C.value_list_t, error) {
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
		case api.Gauge:
			if _, err := C.value_list_add_gauge(ret, C.gauge_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_gauge: %v", err)
			}
		case api.Derive:
			if _, err := C.value_list_add_derive(ret, C.derive_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_derive: %v", err)
			}
		default:
			return nil, fmt.Errorf("not yet supported: %T", v)
		}
	}

	return ret, nil
}

// writer implements the api.Write interface.
type writer struct{}

// NewWriter returns an object implementing the api.Writer interface for the
// collectd daemon.
func NewWriter() api.Writer {
	return writer{}
}

// Write implements the api.Writer interface for the collectd daemon.
func (writer) Write(vl api.ValueList) error {
	return Write(vl)
}

// Write converts a ValueList and calls the plugin_dispatch_values() function
// of the collectd daemon.
func Write(vl api.ValueList) error {
	vlt, err := newValueListT(vl)
	if err != nil {
		return err
	}
	defer C.free(unsafe.Pointer(vlt.values))

	status, err := C.dispatch_values_wrapper(vlt)
	if err != nil {
		return err
	} else if status != 0 {
		return fmt.Errorf("dispatch_values failed with status %d", status)
	}

	return nil
}

type readFunction struct {
	name     string
	callback Reader
}

// RegisterRead registers a new read function with the daemon which is called
// periodically.
func RegisterRead(name string, r Reader) error {
	rf := readFunction{
		name:     name,
		callback: r,
	}

	cGroup := C.CString("golang")
	defer C.free(unsafe.Pointer(cGroup))

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ud := C.user_data_t{
		data:      unsafe.Pointer(&rf),
		free_func: nil,
	}

	status, err := C.register_read_wrapper(cGroup, cName,
		C.plugin_read_cb(C.wrap_read_callback),
		C.cdtime_t(0),
		&ud)
	if err != nil {
		return err
	} else if status != 0 {
		return fmt.Errorf("register_read_wrapper failed with status %d", status)
	}

	return nil
}

//export wrap_read_callback
func wrap_read_callback(ud *C.user_data_t) C.int {
	rf := (*readFunction)(ud.data)

	if err := rf.callback.Read(); err != nil {
		Errorf("%s plugin: Read() failed: %v", rf.name, err)
		return -1
	}

	return 0
}

type writeFunction struct {
	name     string
	callback api.Writer
}

// RegisterWrite registers a new write function with the daemon which is called
// for every metric collected by collectd.
//
// Please note that multiple threads may call this function concurrently. If
// you're accessing shared resources, such as a memory buffer, you have to
// implement appropriate locking around these accesses.
func RegisterWrite(name string, w api.Writer) error {
	wf := writeFunction{
		name:     name,
		callback: w,
	}

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ud := C.user_data_t{
		data:      unsafe.Pointer(&wf),
		free_func: nil,
	}

	status, err := C.register_write_wrapper(cName, C.plugin_write_cb(C.wrap_write_callback), &ud)
	if err != nil {
		return err
	} else if status != 0 {
		return fmt.Errorf("register_write_wrapper failed with status %d", status)
	}

	return nil
}

//export wrap_write_callback
func wrap_write_callback(ds *C.data_set_t, cvl *C.value_list_t, ud *C.user_data_t) C.int {
	wf := (*writeFunction)(ud.data)

	vl := api.ValueList{
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

	for i := C.size_t(0); i < ds.ds_num; i++ {
		dsrc := C.ds_dsrc(ds, i)

		switch dsrc._type {
		case C.DS_TYPE_GAUGE:
			v := C.value_list_get_gauge(cvl, i)
			vl.Values = append(vl.Values, api.Gauge(v))
		case C.DS_TYPE_DERIVE:
			v := C.value_list_get_derive(cvl, i)
			vl.Values = append(vl.Values, api.Derive(v))
		default:
			Errorf("%s plugin: data source type %d is not supported", wf.name, dsrc._type)
			return -1
		}

		vl.DSNames = append(vl.DSNames, C.GoString(&dsrc.name[0]))
	}

	if err := wf.callback.Write(vl); err != nil {
		Errorf("%s plugin: Write() failed: %v", wf.name, err)
		return -1
	}

	return 0
}

//export module_register
func module_register() {
}
