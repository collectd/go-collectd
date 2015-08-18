package cdtime // import "collectd.org/cdtime"

import (
	"testing"
	"time"
)

// TestConversion converts a time.Time to a cdtime.Time and back, expecting the
// original time.Time back.
func TestConversion(t *testing.T) {
	cases := []string{
		"2009-02-04T21:00:57-08:00",
		"2009-02-04T21:00:57.1-08:00",
		"2009-02-04T21:00:57.01-08:00",
		"2009-02-04T21:00:57.001-08:00",
		"2009-02-04T21:00:57.0001-08:00",
		"2009-02-04T21:00:57.00001-08:00",
		"2009-02-04T21:00:57.000001-08:00",
		"2009-02-04T21:00:57.0000001-08:00",
		"2009-02-04T21:00:57.00000001-08:00",
		"2009-02-04T21:00:57.000000001-08:00",
	}

	for _, s := range cases {
		want, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			t.Errorf("time.Parse(%q): got (%v, %v), want (<time.Time>, nil)", s, want, err)
			continue
		}

		cdtime := New(want)
		got := cdtime.Time()
		if !got.Equal(want) {
			t.Errorf("cdtime.Time(): got %v, want %v", got, want)
		}
	}
}
