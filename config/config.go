package config

import (
	"fmt"
	"reflect"
)

type ValueType int

const (
	stringType ValueType = iota
	numberType
	booleanType
)

func (cvt ValueType) String() string {
	return [3]string{"StringValue", "Number", "Boolean"}[cvt]
}

// Value may be either a string, float64 or boolean value.
// This is the Go equivalent of the C type "oconfig_value_t".
type Value struct {
	typ ValueType
	s   string
	f   float64
	b   bool
}

func StringValue(v string) Value   { return Value{typ: stringType, s: v} }
func Float64Value(v float64) Value { return Value{typ: numberType, f: v} }
func BoolValue(v bool) Value       { return Value{typ: booleanType, b: v} }

func (cv Value) String() (string, bool) {
	return cv.s, cv.typ == stringType
}

func (cv Value) Number() (float64, bool) {
	return cv.f, cv.typ == numberType
}

func (cv Value) Boolean() (bool, bool) {
	return cv.b, cv.typ == booleanType
}

// Interface returns the specific value of Value without specifying its type, useful for functions like fmt.Printf
// which can use variables with unknown types.
func (cv Value) Interface() interface{} {
	switch cv.typ {
	case stringType:
		return cv.s
	case numberType:
		return cv.f
	case booleanType:
		return cv.b
	}
	return nil
}

func (cv Value) unmarshal(v reflect.Value) error {
	rvt := v.Type()
	var cvt reflect.Type
	var cvv reflect.Value

	switch cv.typ {
	case stringType:
		cvt = reflect.TypeOf(cv.s)
		cvv = reflect.ValueOf(cv.s)
	case booleanType:
		cvt = reflect.TypeOf(cv.b)
		cvv = reflect.ValueOf(cv.b)
	case numberType:
		cvt = reflect.TypeOf(cv.f)
		cvv = reflect.ValueOf(cv.f)
	default:
		panic("received Value with unknown type")
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

// Block represents one configuration block, which may contain other configuration blocks.
type Block struct {
	Key      string
	Values   []Value
	Children []Block
}

// Unmarshal applies the configuration from a Block to an arbitrary struct.
func (c *Block) Unmarshal(v interface{}) error {
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

	// If config block has child blocks we can only unmarshal to a struct or slice of structs
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
					return fmt.Errorf("in child config block %s: %s", child.Key, err)
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
				if err := cv.unmarshal(tv); err != nil {
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
		return c.Values[0].unmarshal(drv)
	default:
		return fmt.Errorf("cannot unmarshal into type %s", drv.Type())
	}
}

func storeStructConfigValues(cvs []Value, v reflect.Value) error {
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
		if err := cv.unmarshal(args); err != nil {
			return fmt.Errorf("while attempting to unmarshal config value \"%v\" in Args: %s", cv.Interface(), err)
		}
	}
	return nil
}

type Unmarshaler interface {
	UnmarshalConfig(v interface{}) error
}
