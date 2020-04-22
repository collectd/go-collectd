package exec // import "collectd.org/exec"

import (
	"context"
	"os"
	"testing"
	"time"

	"collectd.org/api"
)

func TestSanitizeInterval(t *testing.T) {
	cases := []struct {
		arg  time.Duration
		env  string
		want time.Duration
	}{
		{42 * time.Second, "", 42 * time.Second},
		{42 * time.Second, "23", 42 * time.Second},
		{0, "23", 23 * time.Second},
		{0, "8.15", 8150 * time.Millisecond},
		{0, "", 10 * time.Second},
		{0, "--- INVALID ---", 10 * time.Second},
	}

	for _, tc := range cases {
		if tc.env != "" {
			if err := os.Setenv("COLLECTD_INTERVAL", tc.env); err != nil {
				t.Fatal(err)
			}
		} else { // tc.env == ""
			if err := os.Unsetenv("COLLECTD_INTERVAL"); err != nil {
				t.Fatal(err)
			}
		}

		got := sanitizeInterval(tc.arg)
		if got != tc.want {
			t.Errorf("COLLECTD_INTERVAL=%q sanitizeInterval(%v) = %v, want %v", tc.env, tc.arg, got, tc.want)
		}
	}
}

func Example() {
	e := NewExecutor()

	// simple "value" callback
	answer := func() api.Value {
		return api.Gauge(42)
	}
	e.ValueCallback(answer, &api.ValueList{
		Identifier: api.Identifier{
			Host:         "example.com",
			Plugin:       "golang",
			Type:         "answer",
			TypeInstance: "live_universe_and_everything",
		},
		Interval: time.Second,
	})

	// "complex" void callback
	bicycles := func(ctx context.Context, interval time.Duration) {
		vl := &api.ValueList{
			Identifier: api.Identifier{
				Host:   "example.com",
				Plugin: "golang",
				Type:   "bicycles",
			},
			Interval: interval,
			Time:     time.Now(),
			Values:   make([]api.Value, 1),
		}

		data := []struct {
			TypeInstance string
			Value        api.Gauge
		}{
			{"beijing", api.Gauge(9000000)},
		}
		for _, d := range data {
			vl.Values[0] = d.Value
			vl.Identifier.TypeInstance = d.TypeInstance
			Putval.Write(ctx, vl)
		}
	}
	e.VoidCallback(bicycles, time.Second)

	// blocks forever
	e.Run(context.Background())
}
