package export // import "collectd.org/export"

import (
	"context"
	"errors"
	"expvar"
	"sync"
	"testing"
	"time"

	"collectd.org/api"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestDerive(t *testing.T) {
	// clean up shared resource after testing
	defer func() {
		vars = nil
	}()

	d := NewDeriveString("example.com/TestDerive/derive")
	for i := 0; i < 10; i++ {
		d.Add(i)
	}

	want := &api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "TestDerive",
			Type:   "derive",
		},
		Values: []api.Value{api.Derive(45)},
	}
	got := d.ValueList()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Derive.ValueList() differs (+got/-want):\n%s", diff)
	}

	s := expvar.Get("example.com/TestDerive/derive").String()
	if s != "45" {
		t.Errorf("got %q, want %q", s, "45")
	}
}

func TestGauge(t *testing.T) {
	// clean up shared resource after testing
	defer func() {
		vars = nil
	}()

	g := NewGaugeString("example.com/TestGauge/gauge")
	g.Set(42.0)

	want := &api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "TestGauge",
			Type:   "gauge",
		},
		Values: []api.Value{api.Gauge(42)},
	}
	got := g.ValueList()

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Gauge.ValueList() differs (+got/-want):\n%s", diff)
	}

	s := expvar.Get("example.com/TestGauge/gauge").String()
	if s != "42" {
		t.Errorf("got %q, want %q", s, "42")
	}
}

type testWriter struct {
	got  []*api.ValueList
	done chan<- struct{}
	once *sync.Once
}

func (w *testWriter) Write(ctx context.Context, vl *api.ValueList) error {
	w.got = append(w.got, vl)
	w.once.Do(func() {
		close(w.done)
	})
	return nil
}

func TestRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// clean up shared resource after testing
	defer func() {
		vars = nil
	}()

	d := NewDeriveString("example.com/TestRun/derive")
	d.Add(23)

	g := NewGaugeString("example.com/TestRun/gauge")
	g.Set(42)

	var (
		done = make(chan struct{})
		once sync.Once
	)

	w := testWriter{
		done: done,
		once: &once,
	}

	go func() {
		// when one metric has been written, cancel the context
		<-done
		cancel()
	}()

	err := Run(ctx, &w, Options{Interval: 100 * time.Millisecond})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Run() = %v, want %v", err, context.Canceled)
	}

	want := []*api.ValueList{
		{
			Identifier: api.Identifier{
				Host:   "example.com",
				Plugin: "TestRun",
				Type:   "gauge",
			},
			Time:     time.Now(),
			Interval: 100 * time.Millisecond,
			Values:   []api.Value{api.Gauge(42)},
		},
		{
			Identifier: api.Identifier{
				Host:   "example.com",
				Plugin: "TestRun",
				Type:   "derive",
			},
			Time:     time.Now(),
			Interval: 100 * time.Millisecond,
			Values:   []api.Value{api.Derive(23)},
		},
	}

	ignoreOrder := cmpopts.SortSlices(func(a, b *api.ValueList) bool {
		return a.Identifier.String() < b.Identifier.String()
	})
	approximateTime := cmp.Comparer(func(t0, t1 time.Time) bool {
		diff := t0.Sub(t1)
		if t1.After(t0) {
			diff = t1.Sub(t0)
		}

		return diff < 2*time.Second
	})
	if diff := cmp.Diff(want, w.got, ignoreOrder, approximateTime); diff != "" {
		t.Errorf("received value lists differ (+got/-want):\n%s", diff)
	}
}
