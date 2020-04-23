package cdtime_test

import (
	"encoding/json"
	"testing"
	"time"

	"collectd.org/cdtime"
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

		ct := cdtime.New(want)
		got := ct.Time()
		if !got.Equal(want) {
			t.Errorf("cdtime.Time(): got %v, want %v", got, want)
		}
	}
}

func TestMarshalJSON(t *testing.T) {
	tm := time.Unix(1587671455, 499000000)

	orig := cdtime.New(tm)
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}

	var got cdtime.Time
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	// JSON Marshaling is not loss-less, because it only encodes
	// millisecond precision.
	if got, want := got.String(), "1587671455.499"; got != want {
		t.Errorf("json.Unmarshal() result differs: got %q, want %q", got, want)
	}
}

func TestNewDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want cdtime.Time
	}{
		// 1439981652801860766 * 2^30 / 10^9 = 1546168526406004689.4
		{1439981652801860766 * time.Nanosecond, cdtime.Time(1546168526406004689)},
		// 1439981836985281914 * 2^30 / 10^9 = 1546168724171447263.4
		{1439981836985281914 * time.Nanosecond, cdtime.Time(1546168724171447263)},
		// 1439981880053705608 * 2^30 / 10^9 = 1546168770415815077.4
		{1439981880053705608 * time.Nanosecond, cdtime.Time(1546168770415815077)},
		// 1439981880053705920 * 2^30 / 10^9 = 1546168770415815412.5
		{1439981880053705920 * time.Nanosecond, cdtime.Time(1546168770415815413)},
	}

	for _, tc := range cases {
		d := cdtime.NewDuration(tc.d)
		if got, want := d, tc.want; got != want {
			t.Errorf("NewDuration(%v) = %d, want %d", tc.d, got, want)
		}

		if got, want := d.Duration(), tc.d; got != want {
			t.Errorf("%#v.Duration() = %v, want %v", d, got, want)
		}
	}
}
