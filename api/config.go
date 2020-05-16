package api

import (
	"fmt"
	"reflect"
)

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

func (cv ConfigValue) String() (string, bool) {
	return cv.s, cv.typ == configTypeString
}

func (cv ConfigValue) Number() (float64, bool) {
	return cv.f, cv.typ == configTypeNumber
}

func (cv ConfigValue) Boolean() (bool, bool) {
	return cv.b, cv.typ == configTypeBoolean
}

// Interface returns the specific value of ConfigValue without specifying its type, useful for functions like fmt.Printf
// which can use variables with unknown types.
func (cv ConfigValue) Interface() interface{} {
	switch cv.typ {
	case configTypeString:
		return cv.s
	case configTypeNumber:
		return cv.f
	case configTypeBoolean:
		return cv.b
	}
	return nil
}

func (cv ConfigValue) unmarshalConfig(v reflect.Value) error {
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

// Config represents one configuration block, which may contain other configuration blocks.
type Config struct {
	Key      string
	Values   []ConfigValue
	Children []Config
}

// Unmarshal applies the configuration from a Config to an arbitrary struct.
func (c *Config) Unmarshal(v interface{}) error {
	// If the target supports unmarshalling let it
	if u, ok := v.(Unmarshaler); ok {
		return u.UnmarshalConfig(v)
	}

	// Sanity check value of the interface
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can only unmarshal to a non-nil pointer") // TODO: better error message or nil if preferred
	}

	drv := rv.Elem() // get dereferenced value
	drvk := drv.Kind()

	// If config has child configs we can only unmarshal to a struct or slice of structs
	if len(c.Children) > 0 {
		if drvk != reflect.Struct && (drvk != reflect.Slice || drv.Type().Elem().Kind() != reflect.Struct) {
			return fmt.Errorf("cannot unmarshal a config with children except to a struct or slice of structs")
		}
	}

	switch drvk {
	case reflect.Struct:
		// Unmarshal values from config
		if err := storeStructConfigValues(c.Values, drv); err != nil {
			return fmt.Errorf("while unmarshalling config block values into %s: %s", drv.Type(), err)
		}
		for _, child := range c.Children {
			// If a config has children but the struct has no corresponding field, or the corresponding field is an
			// unexported struct field we throw an error.
			if field := drv.FieldByName(child.Key); field.IsValid() && field.CanInterface() {
				if err := child.Unmarshal(field.Addr().Interface()); err != nil {
					//	if err := child.Unmarshal(field.Interface()); err != nil {
					return fmt.Errorf("in child config %s: %s", child.Key, err)
				}
			} else {
				return fmt.Errorf("found child config block with no corresponding field: %s", child.Key)
			}
		}
		return nil
	case reflect.Slice:
		switch drv.Type().Elem().Kind() {
		case reflect.Struct:
			// Create a temporary Value of the same type as dereferenced value, then get a Value of the same type as
			// its elements. Unmarshal into that Value and append the temporary Value to the original.
			tv := reflect.New(drv.Type().Elem()).Elem()
			if err := c.Unmarshal(tv.Addr().Interface()); err != nil {
				return fmt.Errorf("unmarshaling into temporary value failed: %s", err)
			}
			drv.Set(reflect.Append(drv, tv))
			return nil
		default:
			for _, cv := range c.Values {
				tv := reflect.New(drv.Type().Elem()).Elem()
				if err := cv.unmarshalConfig(tv); err != nil {
					return fmt.Errorf("while unmarhalling values into %s: %s", drv.Type(), err)
				}
				drv.Set(reflect.Append(drv, tv))
			}
			return nil
		}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		if len(c.Values) != 1 {
			return fmt.Errorf("cannot unmarshal config option with %d values into scalar type %s", len(c.Values), drv.Type())
		}
		return c.Values[0].unmarshalConfig(drv)
	default:
		return fmt.Errorf("cannot unmarshal into type %s", drv.Type())
	}
}

func storeStructConfigValues(cvs []ConfigValue, v reflect.Value) error {
	if len(cvs) == 0 {
		return nil
	}
	args := v.FieldByName("Args")
	if !args.IsValid() {
		return fmt.Errorf("cannot unmarshal values to a struct without an Args field")
	}
	if len(cvs) > 1 && args.Kind() != reflect.Slice {
		return fmt.Errorf("cannot unmarshal config block with multiple values to a struct with non-slice Args field")
	}
	for _, cv := range cvs {
		if err := cv.unmarshalConfig(args); err != nil {
			return fmt.Errorf("while attempting to unmarshal config value \"%v\" in Args: %s", cv.Interface(), err)
		}
	}
	return nil
}

type Unmarshaler interface {
	UnmarshalConfig(v interface{}) error
}
