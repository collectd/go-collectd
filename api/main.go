// Package api defines data types representing core collectd data types.
package api // import "collectd.org/api"

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
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

// ParseIdentifier parses the identifier encoded in s and returns it.
func ParseIdentifier(s string) (Identifier, error) {
	fields := strings.Split(s, "/")
	if len(fields) != 3 {
		return Identifier{}, fmt.Errorf("not a valid identifier: %q", s)
	}

	id := Identifier{
		Host:   fields[0],
		Plugin: fields[1],
		Type:   fields[2],
	}

	if i := strings.Index(id.Plugin, "-"); i != -1 {
		id.PluginInstance = id.Plugin[i+1:]
		id.Plugin = id.Plugin[:i]
	}

	if i := strings.Index(id.Type, "-"); i != -1 {
		id.TypeInstance = id.Type[i+1:]
		id.Type = id.Type[:i]
	}

	return id, nil
}

// ValueList represents one (set of) data point(s) of one metric. It is Go's
// equivalent of the C type value_list_t.
type ValueList struct {
	Identifier
	Time     time.Time
	Interval time.Duration
	Values   []Value
	DSNames  []string
}

// DSName returns the name of the data source at the given index. If vl.DSNames
// is nil, returns "value" if there is a single value and a string
// representation of index otherwise.
func (vl *ValueList) DSName(index int) string {
	if vl.DSNames != nil {
		return vl.DSNames[index]
	} else if len(vl.Values) != 1 {
		return strconv.FormatInt(int64(index), 10)
	}

	return "value"
}

// Writer are objects accepting a ValueList for writing, for example to the
// network.
type Writer interface {
	Write(context.Context, *ValueList) error
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

// Dispatcher implements a multiplexer for Writer, i.e. each ValueList
// written to it is copied and written to each registered Writer.
type Dispatcher struct {
	writers []Writer
}

// Add adds a Writer to the Dispatcher.
func (d *Dispatcher) Add(w Writer) {
	d.writers = append(d.writers, w)
}

// Len returns the number of Writers belonging to the Dispatcher.
func (d *Dispatcher) Len() int {
	return len(d.writers)
}

// Write starts a new Goroutine for each Writer which creates a copy of the
// ValueList and then calls the Writer with the copy. It returns nil
// immediately.
func (d *Dispatcher) Write(ctx context.Context, vl *ValueList) error {
	for _, w := range d.writers {
		go func(w Writer) {
			vlCopy := vl
			vlCopy.Values = make([]Value, len(vl.Values))
			copy(vlCopy.Values, vl.Values)

			if err := w.Write(ctx, vlCopy); err != nil {
				log.Printf("%T.Write(): %v", w, err)
			}
		}(w)
	}
	return nil
}

type configValueType int

const (
	configTypeString configValueType = iota
	configTypeNumber
	configTypeBoolean
)

func (cvt configValueType) String() string {
	return [3]string{"String", "Number", "Boolean"}[cvt]
}

// ConfigValue may be either a string, float64 or boolean value.
// This is the Go equivalent of the C type "oconfig_value_t".
type ConfigValue struct {
	typ configValueType
	s   string
	f   float64
	b   bool
}

func (v ConfigValue) String() (string, bool) {
	return v.s, v.typ == configTypeString
}

func (v ConfigValue) Number() (float64, bool) {
	return v.f, v.typ == configTypeNumber
}

func (v ConfigValue) Boolean() (bool, bool) {
	return v.b, v.typ == configTypeBoolean
}

// Untyped returns the specific value of ConfigValue without specifying its type, useful for functions like fmt.Printf
// which can use variables with unknown types.
func (v ConfigValue) Untyped() interface{} {
	switch v.typ {
	case configTypeString:
		return v.s
	case configTypeNumber:
		return v.f
	case configTypeBoolean:
		return v.b
	}
	return nil
}

// Config represents one configuration block, which may contain other configuration blocks.
type Config struct {
	Key      string
	Values   []ConfigValue
	Children []Config
}

// Merge appends other's children to c's children.
// Returns an error if Key or any Values differ.
func (c *Config) Merge(other *Config) error {
	panic("Not yet implemented")
}

// Unmarshal applies the configuration from a Config to an arbitrary struct.
func (c *Config) Unmarshal(v interface{}) error {
	// Sanity check value of the interface
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can only unmarshal to a non-nil pointer") // TODO: better error message or nil if preferred
	}

	// If the target supports unmarshalling let it
	if u, ok := v.(Unmarshaler); ok {
		return u.UnmarshalConfig(v)
	}

	drv := rv.Elem() // get dereferenced value
	drvk := drv.Kind()

	// If config has child configs we can only unmarshal to a struct or slice of structs
	if len(c.Children) > 0 {
		if drvk != reflect.Struct && (drvk != reflect.Slice || drv.Elem().Kind() != reflect.Struct) {
			return fmt.Errorf("cannot unmarshal a config with children except to a struct or slice of structs")
		}
	}

	switch drvk {
	case reflect.Invalid, reflect.Array, reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.UnsafePointer:
		return fmt.Errorf("cannot unmarshal into type %s", drv.Type())
	case reflect.Struct:
		// Unmarshal values from config
		if err := storeStructConfigValues(c.Values, drv); err != nil {
			return fmt.Errorf("while unmarshalling values into %s: %s", drv.Type(), err)
		}
		for i := range c.Children {
			// If a config has children but the struct has no corresponding field, or the corresponding field is an
			// unexported struct field the child is ignored without notice.
			if field := drv.FieldByName(c.Children[i].Key); field.IsValid() && field.CanInterface() {
				fieldPtr := field.Addr().Interface()

				if err := c.Children[i].Unmarshal(fieldPtr); err != nil {
					return fmt.Errorf("while unmarshalling child config with key %s: %s", c.Children[i].Key, err)
				}

			}
		}
		return nil
	case reflect.Slice:
		switch rv.Elem().Kind() {
		case reflect.Struct:
			// Create a temporary Value of the same type as dereferenced value, then get a Value of the same type as
			// its elements. Unmarshal into that Value and append the temporary Value to the original.
			tv := reflect.New(drv.Type()).Elem()
			if err := storeStructConfigValues(c.Values, tv); err != nil {
				return fmt.Errorf("while unmarshalling values into %s: %s", drv.Type(), err)
			}
			for i := range c.Children {
				// If a config has children but the struct type that is an element of this slice has no corresponding
				// field the child is dropped without notice.
				if field := tv.FieldByName(c.Children[i].Key); field.IsValid() {
					if err := c.Children[i].Unmarshal(field); err != nil {
						return fmt.Errorf("while unmarshalling child config with key %s: %s", c.Children[i].Key, err)
					}
				}
			}
			drv.Set(reflect.Append(drv, tv))
			return nil
		default:
			for i := range c.Values {
				tv := reflect.New(drv.Type().Elem()).Elem()
				if err := storeConfigValue(c.Values[i], tv); err != nil {
					return fmt.Errorf("while unmarhalling values into %s: %s", drv.Type(), err)
				}
				drv.Set(reflect.Append(drv, tv))
			}
			return nil
		}
	default: // Kind is one of the number, bool or string kinds
		if len(c.Values) != 1 {
			return fmt.Errorf("cannot unmarshal config with %d values into scalar type %s", len(c.Values), drv.Type())
		}
		return storeConfigValue(c.Values[0], drv)
	}
}

func storeConfigValue(cv ConfigValue, v reflect.Value) error {
	rvt := v.Type()
	var cvt reflect.Type
	var cvv reflect.Value

	switch cv.typ {
	case configTypeString:
		cvt = reflect.TypeOf(cv.s)
		cvv = reflect.ValueOf(cv.s)
	case configTypeBoolean:
		cvt = reflect.TypeOf(cv.b)
		cvv = reflect.ValueOf(cv.b)
	case configTypeNumber:
		cvt = reflect.TypeOf(cv.f)
		cvv = reflect.ValueOf(cv.f)
	default:
		panic("received ConfigValue with unknown type")
	}

	if cvt.ConvertibleTo(rvt) {
		v.Set(cvv.Convert(rvt))
		return nil
	}
	if v.Kind() == reflect.Slice && cvt.ConvertibleTo(rvt.Elem()) {
		v.Set(reflect.Append(v, cvv.Convert(rvt.Elem())))
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s type config value to type %s", cv.typ, v.Type())
}

func storeStructConfigValues(cv []ConfigValue, v reflect.Value) error {
	args := v.FieldByName("Args")
	if !args.IsValid() {
		return fmt.Errorf("cannot unmarshal values to a struct without an Args field")
	}
	if len(cv) > 1 && args.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal multiple config values to a struct with non-slice Args field")
	}
	for i := range cv {
		if err := storeConfigValue(cv[i], args); err != nil {
			return fmt.Errorf("while attempting to unmarshal config value \"%v\" in Args: %s", cv[i].Untyped(), err)
		}
	}
	return nil
}

type Unmarshaler interface {
	UnmarshalConfig(v interface{}) error
}
