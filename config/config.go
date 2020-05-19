package config

import (
	"fmt"
	"math"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

type valueType int

const (
	stringType valueType = iota
	numberType
	booleanType
)

// Value may be either a string, float64 or boolean value.
// This is the Go equivalent of the C type "oconfig_value_t".
type Value struct {
	typ valueType
	s   string
	f   float64
	b   bool
}

// StringValue returns a new string Value.
func StringValue(v string) Value { return Value{typ: stringType, s: v} }

// Float64Value returns a new string Value.
func Float64Value(v float64) Value { return Value{typ: numberType, f: v} }

// BoolValue returns a new boolean Value.
func BoolValue(v bool) Value { return Value{typ: booleanType, b: v} }

// Values allocates and initializes a []Value slice. "string", "float64", and
// "bool" are mapped directly. "[]byte" is converted to a string. Numeric types
// (except complex numbers) are converted to float64. All other values are
// converted to string using the `%v` format.
func Values(values ...interface{}) []Value {
	var ret []Value
	for _, v := range values {
		if v == nil {
			ret = append(ret, Float64Value(math.NaN()))
			continue
		}

		// check for exact matches first.
		switch v := v.(type) {
		case string:
			ret = append(ret, StringValue(v))
			continue
		case []byte:
			ret = append(ret, StringValue(string(v)))
			continue
		case bool:
			ret = append(ret, BoolValue(v))
			continue
		}

		// Handle numerical types that can be converted to float64:
		var (
			valueType   = reflect.TypeOf(v)
			float64Type = reflect.TypeOf(float64(0))
		)
		if valueType.ConvertibleTo(float64Type) {
			v := reflect.ValueOf(v).Convert(float64Type).Interface().(float64)
			ret = append(ret, Float64Value(v))
			continue
		}

		// Last resort: convert to a string using the "fmt" package:
		ret = append(ret, StringValue(fmt.Sprintf("%v", v)))
	}
	return ret
}

// GoString returns a Go statement for creating cv.
func (cv Value) GoString() string {
	switch cv.typ {
	case stringType:
		return fmt.Sprintf("config.StringValue(%q)", cv.s)
	case numberType:
		return fmt.Sprintf("config.Float64Value(%v)", cv.f)
	case booleanType:
		return fmt.Sprintf("config.BoolValue(%v)", cv.b)
	}
	return "<invalid config.Value>"
}

// IsString returns true if cv is a string Value.
func (cv Value) IsString() bool {
	return cv.typ == stringType
}

// String returns Value as a string. Non-string values are formatted according to their default format.
func (cv Value) String() string {
	return fmt.Sprintf("%v", cv.Interface())
}

// Number returns the value of a number Value.
func (cv Value) Number() (float64, bool) {
	return cv.f, cv.typ == numberType
}

// Boolean returns the value of a bool Value.
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
		return fmt.Errorf("unexpected Value type: %v", cv.typ)
	}

	if cvt.ConvertibleTo(rvt) {
		v.Set(cvv.Convert(rvt))
		return nil
	}
	if v.Kind() == reflect.Slice && cvt.ConvertibleTo(rvt.Elem()) {
		v.Set(reflect.Append(v, cvv.Convert(rvt.Elem())))
		return nil
	}
	return fmt.Errorf("cannot unmarshal a %T to a %s", cv.Interface(), v.Type())
}

// Block represents one configuration block, which may contain other configuration blocks.
type Block struct {
	Key      string
	Values   []Value
	Children []Block
}

// Merge appends other's Children to b's Children. If Key or Values differ, an
// error is returned.
func (b *Block) Merge(other Block) error {
	// If b is the zero value, we set it to other.
	if b.Key == "" && b.Values == nil && b.Children == nil {
		*b = other
		return nil
	}

	if b.Key != other.Key || !cmp.Equal(b.Values, other.Values, cmp.AllowUnexported(Value{})) {
		return fmt.Errorf("blocks differ: got {key:%v values:%v}, want {key:%v, values:%v}",
			other.Key, other.Values, b.Key, b.Values)
	}

	b.Children = append(b.Children, other.Children...)
	return nil
}

// Unmarshal applies the configuration from a Block to an arbitrary struct.
func (b *Block) Unmarshal(v interface{}) error {
	// If the target supports unmarshalling let it
	if u, ok := v.(Unmarshaler); ok {
		return u.UnmarshalConfig(*b)
	}

	// Sanity check value of the interface
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("can only unmarshal to a non-nil pointer") // TODO: better error message or nil if preferred
	}

	drv := rv.Elem() // get dereferenced value
	drvk := drv.Kind()

	// If config block has child blocks we can only unmarshal to a struct or slice of structs
	if len(b.Children) > 0 {
		if drvk != reflect.Struct && (drvk != reflect.Slice || drv.Type().Elem().Kind() != reflect.Struct) {
			return fmt.Errorf("cannot unmarshal a config with children except to a struct or slice of structs")
		}
	}

	switch drvk {
	case reflect.Struct:
		// Unmarshal values from config
		if err := storeStructConfigValues(b.Values, drv); err != nil {
			return fmt.Errorf("while unmarshalling config block values into %s: %s", drv.Type(), err)
		}
		for _, child := range b.Children {
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
			if err := b.Unmarshal(tv.Addr().Interface()); err != nil {
				return fmt.Errorf("unmarshaling into temporary value failed: %s", err)
			}
			drv.Set(reflect.Append(drv, tv))
			return nil
		default:
			for _, cv := range b.Values {
				tv := reflect.New(drv.Type().Elem()).Elem()
				if err := cv.unmarshal(tv); err != nil {
					return fmt.Errorf("while unmarhalling values into %s: %s", drv.Type(), err)
				}
				drv.Set(reflect.Append(drv, tv))
			}
			return nil
		}
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Float32, reflect.Float64:
		if len(b.Values) != 1 {
			return fmt.Errorf("cannot unmarshal config option with %d values into scalar type %s", len(b.Values), drv.Type())
		}
		return b.Values[0].unmarshal(drv)
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

// Unmarshaler is the interface implemented by types that can unmarshal a Block representation of themselves.
type Unmarshaler interface {
	UnmarshalConfig(Block) error
}
