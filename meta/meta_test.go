package meta_test

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"sort"
	"testing"
	"time"

	"collectd.org/meta"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func ExampleData() {
	// Allocate new meta.Data object.
	m := meta.Data{
		// Add interger named "answer":
		"answer": meta.Int64(42),
		// Add bool named "panic":
		"panic": meta.Bool(false),
	}

	// Add string named "required":
	m["required"] = meta.String("towel")

	// Remove the "panic" value:
	delete(m, "panic")
}

func ExampleData_exists() {
	m := meta.Data{
		"answer":   meta.Int64(42),
		"panic":    meta.Bool(false),
		"required": meta.String("towel"),
	}

	for _, k := range []string{"answer", "question"} {
		_, ok := m[k]
		fmt.Println(k, "exists:", ok)
	}

	// Output:
	// answer exists: true
	// question exists: false
}

// This example demonstrates how to get a list of keys from meta.Data.
func ExampleData_keys() {
	m := meta.Data{
		"answer":   meta.Int64(42),
		"panic":    meta.Bool(false),
		"required": meta.String("towel"),
	}

	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println(keys)

	// Output:
	// [answer panic required]
}

func ExampleEntry() {
	// Allocate an int64 Entry.
	answer := meta.Int64(42)

	// Read back the "answer" value and ensure it is in fact an int64.
	a, ok := answer.Int64()
	if !ok {
		log.Fatal("Answer is not an int64")
	}
	fmt.Printf("The answer is between %d and %d\n", a-1, a+1)

	// Allocate a string Entry.
	required := meta.String("towel")

	// String is a bit different, because Entry.String() does not return a boolean.
	// Check that "required" is a string and read it into a variable.
	if !required.IsString() {
		log.Fatal("required is not a string")
	}
	fmt.Println("You need a " + required.String())

	// The fmt.Stringer interface is implemented for all value types. To
	// print a string with default formatting, rely on the String() method:
	p := meta.Bool(false)
	fmt.Printf("Should I panic? %v\n", p)

	// Output:
	// The answer is between 41 and 43
	// You need a towel
	// Should I panic? false
}

func ExampleEntry_Interface() {
	rand.Seed(time.Now().UnixNano())
	m := meta.Data{}

	// Create a value with unknown type. "key" is either a string,
	// or an int64.
	switch rand.Intn(2) {
	case 0:
		m["key"] = meta.String("value")
	case 1:
		m["key"] = meta.Int64(42)
	}

	// Scenario 0: A specific type is expected. Report an error that
	// includes the actual type in the error message, if the value is of a
	// different type.
	if _, ok := m["key"].Int64(); !ok {
		err := fmt.Errorf("key is a %T, want an int64", m["key"].Interface())
		fmt.Println(err) // prints "key is a string, want an int64"
	}

	// Scenario 1: Multiple or all types need to be handled, for example to
	// encode the meta data values. The most elegant solution for that is a
	// type switch.
	switch v := m["key"].Interface().(type) {
	case string:
		// string-specific code here
	case int64:
		// The above code skipped printing this, so print it here so
		// this example produces the same output every time, despite
		// the randomness.
		fmt.Println("key is a string, want an int64")
	default:
		// Report the actual type if "key" is an unexpected type.
		err := fmt.Errorf("unexpected type %T", v)
		log.Fatal(err)
	}

	// Output:
	// key is a string, want an int64
}

func TestMarshalJSON(t *testing.T) {
	cases := []struct {
		d    meta.Data
		want string
	}{
		{meta.Data{"foo": meta.Bool(true)}, `{"foo":true}`},
		{meta.Data{"foo": meta.Float64(20.0 / 3.0)}, `{"foo":6.666666666666667}`},
		{meta.Data{"foo": meta.Float64(math.NaN())}, `{"foo":null}`},
		{meta.Data{"foo": meta.Int64(-42)}, `{"foo":-42}`},
		{meta.Data{"foo": meta.UInt64(42)}, `{"foo":42}`},
		{meta.Data{"foo": meta.String(`Hello "World"!`)}, `{"foo":"Hello \"World\"!"}`},
		{meta.Data{"foo": meta.Entry{}}, `{"foo":null}`},
	}

	for _, tc := range cases {
		got, err := json.Marshal(tc.d)
		if err != nil {
			t.Errorf("json.Marshal(%#v) = %v", tc.d, err)
			continue
		}

		if diff := cmp.Diff(tc.want, string(got)); diff != "" {
			t.Errorf("json.Marshal(%#v) differs (+got/-want):\n%s", tc.d, diff)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {
	cases := []struct {
		in      string
		want    meta.Data
		wantErr bool
	}{
		{
			in:   `{}`,
			want: meta.Data{},
		},
		{
			in:   `{"bool":true}`,
			want: meta.Data{"bool": meta.Bool(true)},
		},
		{
			in:   `{"string":"bar"}`,
			want: meta.Data{"string": meta.String("bar")},
		},
		{
			in:   `{"int":42}`,
			want: meta.Data{"int": meta.Int64(42)},
		},
		{ // 9223372036854777144 exceeds 2^63-1
			in:   `{"uint":9223372036854777144}`,
			want: meta.Data{"uint": meta.UInt64(9223372036854777144)},
		},
		{
			in:   `{"float":42.25}`,
			want: meta.Data{"float": meta.Float64(42.25)},
		},
		{
			in:   `{"float":null}`,
			want: meta.Data{"float": meta.Float64(math.NaN())},
		},
		{
			in: `{"bool":false,"string":"","int":-9223372036854775808,"uint":18446744073709551615,"float":0.00006103515625}`,
			want: meta.Data{
				"bool":   meta.Bool(false),
				"string": meta.String(""),
				"int":    meta.Int64(-9223372036854775808),
				"uint":   meta.UInt64(18446744073709551615),
				"float":  meta.Float64(0.00006103515625),
			},
		},
		{
			in:      `{"float":["invalid", "type"]}`,
			wantErr: true,
		},
	}

	for _, c := range cases {
		var got meta.Data
		err := json.Unmarshal([]byte(c.in), &got)
		if gotErr := err != nil; gotErr != c.wantErr {
			t.Errorf("Unmarshal() = %v, want error: %v", err, c.wantErr)
		}
		if err != nil || c.wantErr {
			continue
		}

		opts := []cmp.Option{
			cmp.AllowUnexported(meta.Entry{}),
			cmpopts.EquateNaNs(),
		}
		if diff := cmp.Diff(c.want, got, opts...); diff != "" {
			t.Errorf("Unmarshal() result differs (+got/-want):\n%s", diff)
		}
	}
}

func TestEntry(t *testing.T) {
	cases := []struct {
		typ         string
		e           meta.Entry
		wantBool    bool
		wantFloat64 bool
		wantInt64   bool
		wantUInt64  bool
		wantString  bool
		s           string
	}{
		{
			typ:      "bool",
			e:        meta.Bool(true),
			wantBool: true,
			s:        "true",
		},
		{
			typ:         "float64",
			e:           meta.Float64(20.0 / 3.0),
			wantFloat64: true,
			s:           "6.66666666666667",
		},
		{
			typ:       "int64",
			e:         meta.Int64(-9223372036854775808),
			wantInt64: true,
			s:         "-9223372036854775808",
		},
		{
			typ:        "uint64",
			e:          meta.UInt64(18446744073709551615),
			wantUInt64: true,
			s:          "18446744073709551615",
		},
		{
			typ:        "string",
			e:          meta.String("Hello, World!"),
			wantString: true,
			s:          "Hello, World!",
		},
		{
			// meta.Entry's zero value
			typ: "<nil>",
			s:   "<nil>",
		},
	}

	for _, tc := range cases {
		if v, got := tc.e.Bool(); got != tc.wantBool {
			t.Errorf("%#v.Bool() = (%v, %v), want (_, %v)", tc.e, v, got, tc.wantBool)
		}

		if v, got := tc.e.Float64(); got != tc.wantFloat64 {
			t.Errorf("%#v.Float64() = (%v, %v), want (_, %v)", tc.e, v, got, tc.wantFloat64)
		}

		if v, got := tc.e.Int64(); got != tc.wantInt64 {
			t.Errorf("%#v.Int64() = (%v, %v), want (_, %v)", tc.e, v, got, tc.wantInt64)
		}

		if v, got := tc.e.UInt64(); got != tc.wantUInt64 {
			t.Errorf("%#v.UInt64() = (%v, %v), want (_, %v)", tc.e, v, got, tc.wantUInt64)
		}

		if got := tc.e.IsString(); got != tc.wantString {
			t.Errorf("%#v.IsString() = %v, want %v", tc.e, got, tc.wantString)
		}

		if got, want := tc.e.String(), tc.s; got != want {
			t.Errorf("%#v.String() = %q, want %q", tc.e, got, want)
		}

		if got, want := fmt.Sprintf("%T", tc.e.Interface()), tc.typ; got != want {
			t.Errorf("%#v.Interface() = type %s, want type %s", tc.e, got, want)
		}
	}
}

func TestData_Clone(t *testing.T) {
	want := meta.Data{
		"bool":   meta.Bool(false),
		"string": meta.String(""),
		"int":    meta.Int64(-9223372036854775808),
		"uint":   meta.UInt64(18446744073709551615),
		"float":  meta.Float64(0.00006103515625),
	}

	got := want.Clone()

	opts := []cmp.Option{
		cmp.AllowUnexported(meta.Entry{}),
		cmpopts.EquateNaNs(),
	}
	if diff := cmp.Diff(want, got, opts...); diff != "" {
		t.Errorf("Data.Clone() contains differences (+got/-want):\n%s", diff)
	}

	want = nil
	if got := meta.Data(nil).Clone(); got != nil {
		t.Errorf("Data(nil).Clone() = %v, want %v", got, nil)
	}
}
