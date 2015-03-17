package api

import (
	"reflect"
	"testing"
	"time"
)

func TestValueList(t *testing.T) {
	vlWant := ValueList{
		Identifier: Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Time:     time.Unix(1426585562, 999000000),
		Interval: 10 * time.Second,
		Values:   []Value{Gauge(42)},
	}

	want := `{"values":[42],"dstypes":["gauge"],"dsnames":["value"],"time":1426585562.999,"interval":10.000,"host":"example.com","plugin":"golang","type":"gauge"}`

	got, err := vlWant.MarshalJSON()
	if err != nil || string(got) != want {
		t.Errorf("got (%s, %v), want (%s, nil)", got, err, want)
	}

	var vlGot ValueList
	if err := vlGot.UnmarshalJSON([]byte(want)); err != nil {
		t.Errorf("got %v, want nil)", err)
	}

	// Conversion to float64 and back takes its toll -- the conversion is
	// very accurate, but not bit-perfect.
	vlGot.Time = vlGot.Time.Round(time.Millisecond)
	if !reflect.DeepEqual(vlWant, vlGot) {
		t.Errorf("got %#v, want %#v)", vlGot, vlWant)
	}
}
