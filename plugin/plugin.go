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
	  "time"

	  "collectd.org/api"
	  "collectd.org/plugin"
  )

  type ExamplePlugin struct{}

  func (*ExamplePlugin) Read() error {
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
	  if err := plugin.Write(context.Background(), vl); err != nil {
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
// void value_list_add_counter (value_list_t *, counter_t);
// void value_list_add_derive  (value_list_t *, derive_t);
// void value_list_add_gauge   (value_list_t *, gauge_t);
// counter_t value_list_get_counter (value_list_t *, size_t);
// derive_t  value_list_get_derive  (value_list_t *, size_t);
// gauge_t   value_list_get_gauge   (value_list_t *, size_t);
//
// typedef int (*plugin_complex_config_cb)(oconfig_item_t);
//
// int wrap_read_callback(user_data_t *);
//
// int register_write_wrapper (char const *, plugin_write_cb, user_data_t *);
// int wrap_write_callback(data_set_t *, value_list_t *, user_data_t *);
//
// int register_shutdown_wrapper (char *, plugin_shutdown_cb);
// int wrap_shutdown_callback(void);
//
// int register_complex_config_wrapper (char *, plugin_complex_config_cb);
// int process_complex_config(oconfig_item_t*);
// char *go_get_string_value(oconfig_item_t*, int);
// double go_get_number_value(oconfig_item_t*, int);
// int go_get_boolean_value(oconfig_item_t*, int);
// int go_get_value_type(oconfig_item_t*, int);
import "C"

import (
	"collectd.org/api"
	"collectd.org/cdtime"
	"context"
	"fmt"
	"reflect"
	"unsafe"
)

var (
	ctx = context.Background()
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
				return nil, fmt.Errorf("value_list_add_counter: %v", err)
			}
		case api.Derive:
			if _, err := C.value_list_add_derive(ret, C.derive_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_derive: %v", err)
			}
		case api.Gauge:
			if _, err := C.value_list_add_gauge(ret, C.gauge_t(v)); err != nil {
				return nil, fmt.Errorf("value_list_add_gauge: %v", err)
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
func (writer) Write(_ context.Context, vl *api.ValueList) error {
	return Write(vl)
}

// Write converts a ValueList and calls the plugin_dispatch_values() function
// of the collectd daemon.
func Write(vl *api.ValueList) error {
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

	readFuncs[name] = r
	return nil
}

//export wrap_read_callback
func wrap_read_callback(ud *C.user_data_t) C.int {
	name := C.GoString((*C.char)(ud.data))
	r, ok := readFuncs[name]
	if !ok {
		return -1
	}

	if err := r.Read(); err != nil {
		Errorf("%s plugin: Read() failed: %v", name, err)
		return -1
	}

	return 0
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
		free_func: nil,
	}

	status, err := C.register_write_wrapper(cName, C.plugin_write_cb(C.wrap_write_callback), &ud)
	if err != nil {
		return err
	} else if status != 0 {
		return fmt.Errorf("register_write_wrapper failed with status %d", status)
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

	if err := w.Write(ctx, vl); err != nil {
		Errorf("%s plugin: Write() failed: %v", name, err)
		return -1
	}

	return 0
}

// First declare some types, interfaces, general functions

// Shutters are objects that when called will shut down the plugin gracefully
type Shutter interface {
	Shutdown() error
}

// shutdownFuncs holds references to all shutdown callbacks
var shutdownFuncs = make(map[string]Shutter)

//export wrap_shutdown_callback
func wrap_shutdown_callback() C.int {
	fmt.Println("wrap_shutdown_callback called")
	if len(shutdownFuncs) <= 0 {
		return 0
	}
	for n, s := range shutdownFuncs {
		if err := s.Shutdown(); err != nil {
			Errorf("%s plugin: Shutdown() failed: %v", n, s)
			return -1
		}
	}
	return 0
}

// RegisterShutdown registers a shutdown function with the daemon which is called
// when the plugin is required to shutdown gracefully.
func RegisterShutdown(name string, s Shutter) error {
	// Only register a callback the first time one is implemented, subsequent
	// callbacks get added to a list and called at the same time
	if len(shutdownFuncs) <= 0 {
		cName := C.CString(name)
		cCallback := C.plugin_shutdown_cb(C.wrap_shutdown_callback)

		status, err := C.register_shutdown_wrapper(cName, cCallback)
		if err != nil {
			Errorf("register_shutdown_wrapper failed with status: %v", status)
			return err
		}
	}
	shutdownFuncs[name] = s
	return nil
}

// Configuration defines the interface for complex_configure callbacks, i.e. Go
// structs that can hold and validate a configuration
type Configuration interface {
	Validate() error
}

var config_target *Configuration
var config_targets map[string]*Configuration

var configured map[string]chan struct{}

// getOconfigChildren takes a C.oconfig_item_t and returns an array of Go
// pointers to its child items. Getting the child pointers requires a little
// more arithmetic than in C as we can't just increment a pointer to get the
// next array item
func getOconfigChildren(oconfig *C.oconfig_item_t) (output []*C.oconfig_item_t) {
	start := unsafe.Pointer(oconfig.children)
	size := unsafe.Sizeof(*oconfig.children)
	for i := 0; i < int(oconfig.children_num); i++ {
		child := (*C.oconfig_item_t)(unsafe.Pointer(uintptr(start) + size*uintptr(i)))
		output = append(output, child)
	}
	return
}

// checkMatchable returns an error if the provided source cannot be copied to
// the target, or nil if it's OK
func checkMatchable(oconfig *C.oconfig_item_t, target reflect.Value) error {
	if oconfig.children_num > 0 { // if it has children we need a corresponding struct
		if target.Kind() != reflect.Struct {
			return fmt.Errorf("did not have corresponding struct")
		}
	}
	if oconfig.values_num > 1 { // if it has multiple values we need a slice
		if target.Kind() != reflect.Slice {
			return fmt.Errorf("did not have corresponding slice")
		}
	}
	if oconfig.values_num == 1 { // if it has a single value we need either a string, int or bool
		valType, err := C.go_get_value_type(oconfig, C.int(0))
		if err != nil {
			return fmt.Errorf("unable to determine type of value")
		}
		switch valType {
		case C.int(0): //string
			if target.Kind() != reflect.String {
				return fmt.Errorf("found string, plugin expects %v", target.Kind().String())
			}
		case C.int(1): // number/double/float64
			if target.Kind() != reflect.Float64 {
				return fmt.Errorf("found number, plugin expects %v", target.Kind().String())
			}
		case C.int(2): // boolean
			if target.Kind() != reflect.Bool {
				return fmt.Errorf("found boolean, plugin expects %v", target.Kind().String())
			}
		default:
			return fmt.Errorf("unsupported Collectd config item type")
		}
	}
	return nil
}

// checkUsable iterates over a provided reflect.Value and attempts to make sure
// it A) contains only string, bool, or float64 fields, or slices or structs of
// these, and B) contains only settable, exported fields
// Unfortunately as we start with a reflect.Value and no reflect.Type we can't
// actually report the name of the failing field. Fortunately this should be a
// plugin error only (not possible for a user to trigger) and resolved during
// notmal testing.
func checkUsable(target reflect.Value) error {
	switch target.Kind().String() {
	case "struct":
		for i := 0; i < target.NumField(); i++ {
			err := checkUsable(target.Field(i))
			if err != nil {
				return fmt.Errorf("in struct: %v", err)
			}
		}
	case "slice":
		if target.CanSet() != true {
			return fmt.Errorf("unsettable value (unexported field?)")
		}
		// Get an Interface of the slice, extract a field and get the field's type
		switch reflect.TypeOf(target.Interface()).Elem().Kind() {
		case reflect.String:
		case reflect.Float64:
		case reflect.Bool:
		default:
			return fmt.Errorf("unsupported variable type (slices must be of string, float64, or bool)")
		}
	case "string":
		if target.CanSet() != true {
			return fmt.Errorf("unsettable value (unexported field?)")
		}
		return nil
	case "float64":
		if target.CanSet() != true {
			return fmt.Errorf("unsettable value (unexported field?)")
		}
		return nil
	case "bool":
		if target.CanSet() != true {
			return fmt.Errorf("unsettable value (unexported field?)")
		}
		return nil
	default:
		fmt.Println(target.Kind().String())
		return fmt.Errorf("unsupported variable type (must be string, float64, bool, or slice of these, or struct)")
	}
	return nil // here to keep the compiler happy
}

// assignValue copies a oconfig_value to a target Value using the appropriate function
func assignValue(oconfig_item *C.oconfig_item_t, target reflect.Value, index int) error {
	valType, err := C.go_get_value_type(oconfig_item, C.int(index))
	if err != nil {
		return fmt.Errorf("unable to determine type of value")
	}
	switch valType {
	case C.int(0): //string
		target.SetString(C.GoString(C.go_get_string_value(oconfig_item, C.int(index))))
	case C.int(1): // number/double/float64
		target.SetFloat(float64(C.go_get_number_value(oconfig_item, C.int(index))))
	case C.int(2): // boolean
		target.SetBool(C.go_get_number_value(oconfig_item, C.int(index)) == 1)
	default: // Unknown type
		return fmt.Errorf("unsupported config item type")
	}
	return nil
}

// assignSlice generates a slice of the target type, loads in each value and assigns it to the target
func assignSlice(oconfig_item *C.oconfig_item_t, target reflect.Value) error {
	s := reflect.MakeSlice(target.Type(), int(oconfig_item.values_num), int(oconfig_item.values_num))
	for i := 0; i < int(oconfig_item.values_num); i++ {
		assignValue(oconfig_item, s.Index(i), i)
	}
	target.Set(s)
	return nil
}

// assignConfig takes an oconfig_item_t struct and attempts to assign data to corresponding
// fields in the provided config_target. A lack of corresponding field is not an error, a
// corresponding field of the wrong type is.
func assignConfig(oconfig *C.oconfig_item_t, target reflect.Value, isRoot bool) error {
	// root oconfig_item's value is the name of the plugin and unused here
	if isRoot != true {
		if err := checkMatchable(oconfig, target); err != nil {
			return fmt.Errorf("unmatchable value (%v)", err)
		}
	}
	if int(oconfig.values_num) > 1 { // multiple values
		assignSlice(oconfig, target)
	} else if (oconfig.values_num == 1) && (isRoot == false) { // single value
		assignValue(oconfig, target, 0)
	} else if oconfig.children_num > 0 { // a container of oconfig children
		for _, child := range getOconfigChildren(oconfig) {
			childkey := C.GoString(child.key)
			if target.FieldByName(childkey).IsValid() != true { // if there's no corresponding target field, skip it
				fmt.Printf("Ignoring unexpected key in config: %v\n", childkey)
				continue
			}
			if err := assignConfig(child, target.FieldByName(childkey), false); err != nil {
				return fmt.Errorf("in %v: %v\n", childkey, err)
			}
		}
	}
	return nil
}

// RequestConfiguration registers a Configuration struct with the daemon and
// requests a callback to fill it in. It returns a channel which will be
// closed when a configuration has been loaded successfully
func RequestConfiguration(name string, c Configuration) (chan struct{}, error) {
	if err := checkUsable(reflect.ValueOf(c).Elem()); err != nil {
		fmt.Printf("while loading plugin received error: %v\n", err)
		return nil, fmt.Errorf("plugin contains unusable configuration struct")
	}
	if len(config_targets) == 0 {
		config_targets = make(map[string]*Configuration)
		configured = make(map[string]chan struct{})
	}
	config_targets[name] = &c
	configured[name] = make(chan struct{})
	cName := C.CString(name)
	cCallback := C.plugin_complex_config_cb(C.process_complex_config)
	status, err := C.register_complex_config_wrapper(cName, cCallback)
	if err != nil {
		Errorf("register_complex_config_wrapper failed with status %v", status)
		return nil, err
	}
	return configured[name], nil
}

// process_complex_config receives the plugin config from Collectd and attempts to
// initialise the target config. If the configuration is parsed without fatal errors
// it calls the attached Validate() function, which allows the plugin developer to
// confirm that they've received sufficient valid configuration. If that succeeds
// we finally close the Configured channel to indicate that the configuration is now
// assigned and validated.
//export process_complex_config
func process_complex_config(oconfig *C.oconfig_item_t) C.int {
	oname := C.GoString(C.go_get_string_value(oconfig, C.int(0)))
	var ok bool
	if config_target, ok = config_targets[oname]; !ok {
		fmt.Printf("Error: Received configuration for unknown plugin name %v\n", oname)
		return C.int(1)
	}
	target := reflect.ValueOf(*config_target).Elem()
	err := assignConfig(oconfig, target, true)
	if err != nil {
		fmt.Printf("Error: Invalid configuration %v\n", err)
		return C.int(1)
	}
	valArgs := make([]reflect.Value, 0)
	valResp := target.MethodByName("Validate").Call(valArgs)
	if err := valResp[0].Interface(); err != nil {
		fmt.Printf("Error: Plugin-specific validation returned error: %v\n", err)
		return C.int(1)
	}
	close(configured[oname])
	return C.int(0)
}

//export module_register
func module_register() {
}
