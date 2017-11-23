package meta

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
)

func ExampleData() {
	// Allocate new meta.Data object.
	m := make(Data)

	// Add interger named "answer":
	m["answer"] = Int64(42)

	// Add string named "required":
	m["required"] = String("towel")

	// Read back a value:
	answer, ok := m["answer"].Int64()
	if !ok || answer != 42 {
		fmt.Printf("m[%q] = (%d, %v), want (%d, %v)", "answer", answer, ok, 42, true)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	cases := []struct {
		in   string
		want Data
	}{
		{
			in:   `{}`,
			want: nil,
		},
		{
			in:   `{"bool":true}`,
			want: Data{"bool": Bool(true)},
		},
		{
			in:   `{"string":"bar"}`,
			want: Data{"string": String("bar")},
		},
		{
			in:   `{"int":42}`,
			want: Data{"int": Int64(42)},
		},
		{ // 9223372036854777144 exceeds 2^63-1
			in:   `{"uint":9223372036854777144}`,
			want: Data{"uint": UInt64(9223372036854777144)},
		},
		{
			in:   `{"float":42.25}`,
			want: Data{"float": Float64(42.25)},
		},
		{ // 9223372036854777144 exceeds 2^63-1
			in: `{"bool":false,"string":"","int":-9223372036854775808,"uint":18446744073709551615,"float":0.00006103515625}`,
			want: Data{
				"bool":   Bool(false),
				"string": String(""),
				"int":    Int64(-9223372036854775808),
				"uint":   UInt64(18446744073709551615),
				"float":  Float64(0.00006103515625),
			},
		},
	}

	for _, c := range cases {
		var got Data
		if err := json.Unmarshal([]byte(c.in), &got); err != nil {
			t.Errorf("Unmarshal() = %v", err)
			continue
		}

		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("Unmarshal() = %#v, want %#v", got, c.want)
		}
	}
}

// TestUnmarshalJSON_NaN tests that null gets converted to Float64(math.NaN()).
// We cannot add this to the above table test, because reflect.DeepEqual() does
// not compare two NaNs as equal.
func TestUnmarshalJSON_NaN(t *testing.T) {
	var d Data
	if err := json.Unmarshal([]byte(`{"float":null}`), &d); err != nil {
		t.Errorf("Unmarshal() = %v", err)
		return
	}

	got, ok := d["float"].Float64()
	if !ok || !math.IsNaN(got) {
		t.Errorf(`got["float"].Float64() = (%v, %v), want (%v, %v)`, got, ok, math.NaN(), true)
	}
}
