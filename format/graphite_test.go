package format

import (
	"bytes"
	"testing"
	"time"

	"collectd.org/api"
)

func TestWrite(t *testing.T) {
	vl := api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Time:     time.Unix(1426975989, 1),
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Gauge(42)},
	}

	buf := &bytes.Buffer{}

	g := &Graphite{
		W:                 buf,
		Prefix:            "-->",
		Suffix:            "<--",
		EscapeChar:        "_",
		SeparateInstances: false,
		AlwaysAppendDS:    true,
	}

	if err := g.Write(vl); err != nil {
		t.Errorf("got %v, want %v", err, nil)
	}

	want := "-->example_com<--.golang.gauge.value 42 1426975989\r\n"
	got := buf.String()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
