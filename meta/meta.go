// Package meta provides data types for collectd meta data.
//
// Meta data can be associated with value lists (api.ValueList) and
// notifications (not yet implemented in the collectd Go API).
package meta

import (
	"encoding/json"
	"fmt"
	"math"
)

type entryType int

const (
	_ entryType = iota
	metaStringType
	metaInt64Type
	metaUInt64Type
	metaFloat64Type
	metaBoolType
)

// Entry is the interface implemented by all meta data entries. The types
// implementing this interface are String, Int64, UInt64, Float64 and Bool. The
// interface contains private functions to prevent other packages from
// implementing it, so that the use really is limited to the previously listed
// types.
type Entry interface {
	String() (string, bool)
	Int64() (int64, bool)
	UInt64() (uint64, bool)
	Float64() (float64, bool)
	Bool() (bool, bool)

	getType() entryType
}

// Data is a map of meta data values. No setter and getter methods are
// implemented for this, callers are expected to add and remove entries as they
// would from a normal map.
type Data map[string]Entry

func (d Data) Clone() Data {
	if d == nil {
		return nil
	}

	cpy := make(Data)
	for k, v := range d {
		cpy[k] = v
	}
	return cpy
}

// UnmarshalJSON implements the "encoding/json".Unmarshaller interface.
func (d *Data) UnmarshalJSON(raw []byte) error {
	var m map[string]jsonEntry
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}

	for k, v := range m {
		if *d == nil {
			*d = Data{}
		}
		(*d)[k] = v.Entry
	}
	return nil
}

// jsonEntry is a helper type implementing "encoding/json".Unmarshaller for Entry.
type jsonEntry struct {
	Entry
}

func (e *jsonEntry) UnmarshalJSON(raw []byte) error {
	var b *bool
	if json.Unmarshal(raw, &b) == nil && b != nil {
		e.Entry = Bool(*b)
		return nil
	}

	var s *string
	if json.Unmarshal(raw, &s) == nil && s != nil {
		e.Entry = String(*s)
		return nil
	}

	var i *int64
	if json.Unmarshal(raw, &i) == nil && i != nil {
		e.Entry = Int64(*i)
		return nil
	}

	var u *uint64
	if json.Unmarshal(raw, &u) == nil && u != nil {
		e.Entry = UInt64(*u)
		return nil
	}

	var f *float64
	if json.Unmarshal(raw, &f) == nil {
		if f == nil {
			nan := math.NaN()
			f = &nan
		}
		e.Entry = Float64(*f)
		return nil
	}

	return fmt.Errorf("unable to parse %q as meta entry", raw)
}

// String is a string implementing the Entry interface.
type String string

func (s String) String() (string, bool) {
	return string(s), true
}

func (String) Int64() (int64, bool) {
	return 0, false
}

func (String) UInt64() (uint64, bool) {
	return 0, false
}

func (String) Float64() (float64, bool) {
	return 0, false
}

func (String) Bool() (bool, bool) {
	return false, false
}

func (String) getType() entryType {
	return metaStringType
}

// Int64 is a int64 implementing the Entry interface.
type Int64 int64

func (Int64) String() (string, bool) {
	return "", false
}

func (s Int64) Int64() (int64, bool) {
	return int64(s), true
}

func (Int64) UInt64() (uint64, bool) {
	return 0, false
}

func (Int64) Float64() (float64, bool) {
	return 0, false
}

func (Int64) Bool() (bool, bool) {
	return false, false
}

func (Int64) getType() entryType {
	return metaInt64Type
}

// UInt64 is a uint64 implementing the Entry interface.
type UInt64 uint64

func (UInt64) String() (string, bool) {
	return "", false
}

func (UInt64) Int64() (int64, bool) {
	return 0, false
}

func (s UInt64) UInt64() (uint64, bool) {
	return uint64(s), true
}

func (UInt64) Float64() (float64, bool) {
	return 0, false
}

func (UInt64) Bool() (bool, bool) {
	return false, false
}

func (UInt64) getType() entryType {
	return metaUInt64Type
}

// Float64 is a float64 implementing the Entry interface.
type Float64 float64

func (Float64) String() (string, bool) {
	return "", false
}

func (Float64) Int64() (int64, bool) {
	return 0, false
}

func (Float64) UInt64() (uint64, bool) {
	return 0, false
}

func (s Float64) Float64() (float64, bool) {
	return float64(s), true
}

func (Float64) Bool() (bool, bool) {
	return false, false
}

func (Float64) getType() entryType {
	return metaFloat64Type
}

// Bool is a bool implementing the Entry interface.
type Bool bool

func (Bool) String() (string, bool) {
	return "", false
}

func (Bool) Int64() (int64, bool) {
	return 0, false
}

func (Bool) UInt64() (uint64, bool) {
	return 0, false
}

func (Bool) Float64() (float64, bool) {
	return 0, false
}

func (s Bool) Bool() (bool, bool) {
	return bool(s), true
}

func (Bool) getType() entryType {
	return metaBoolType
}
