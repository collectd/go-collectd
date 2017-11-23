package meta

import (
	"encoding/json"
	"log"
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

	// Read back a value where you expect a certain type:
	if v, ok := m["answer"]; ok {
		answer, ok := v.Int64()
		if !ok {
			log.Printf("answer is a %T, expected an Int64", v)
		}
		if answer != 42 {
			log.Printf("answer is a %v, which is obviously wrong", answer)
		}
	}

	// Read back a value where you don't know the type:
	if v, ok := m["required"]; ok {
		switch v := v.(type) {
		case String:
			// here, v is of type String. There are two ways to convert it
			// to non-interface type:
			//   * s := string(v)
			//   * s, ok := v.String()
			msg := "You need a " + string(v)
			log.Print(msg)
		default:
			log.Printf("You need a %v, but I don't know what that is", v)
		}
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
