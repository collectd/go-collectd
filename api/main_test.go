package api_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"collectd.org/api"
	"github.com/google/go-cmp/cmp"
)

func TestParseIdentifier(t *testing.T) {
	cases := []struct {
		Input string
		Want  api.Identifier
	}{
		{
			Input: "example.com/golang/gauge",
			Want: api.Identifier{
				Host:   "example.com",
				Plugin: "golang",
				Type:   "gauge",
			},
		},
		{
			Input: "example.com/golang-foo/gauge-bar",
			Want: api.Identifier{
				Host:           "example.com",
				Plugin:         "golang",
				PluginInstance: "foo",
				Type:           "gauge",
				TypeInstance:   "bar",
			},
		},
		{
			Input: "example.com/golang-a-b/gauge-b-c",
			Want: api.Identifier{
				Host:           "example.com",
				Plugin:         "golang",
				PluginInstance: "a-b",
				Type:           "gauge",
				TypeInstance:   "b-c",
			},
		},
	}

	for i, c := range cases {
		if got, err := api.ParseIdentifier(c.Input); got != c.Want || err != nil {
			t.Errorf("case %d: got (%v, %v), want (%v, %v)", i, got, err, c.Want, nil)
		}
	}

	failures := []string{
		"example.com/golang",
		"example.com/golang/gauge/extra",
	}

	for _, c := range failures {
		if got, err := api.ParseIdentifier(c); err == nil {
			t.Errorf("got (%v, %v), want (%v, !%v)", got, err, api.Identifier{}, nil)
		}
	}
}

func TestIdentifierString(t *testing.T) {
	id := api.Identifier{
		Host:   "example.com",
		Plugin: "golang",
		Type:   "gauge",
	}

	cases := []struct {
		PluginInstance, TypeInstance string
		Want                         string
	}{
		{"", "", "example.com/golang/gauge"},
		{"foo", "", "example.com/golang-foo/gauge"},
		{"", "foo", "example.com/golang/gauge-foo"},
		{"foo", "bar", "example.com/golang-foo/gauge-bar"},
	}

	for _, c := range cases {
		id.PluginInstance = c.PluginInstance
		id.TypeInstance = c.TypeInstance

		got := id.String()
		if got != c.Want {
			t.Errorf("got %q, want %q", got, c.Want)
		}
	}
}

type testWriter struct {
	got *api.ValueList
	wg  *sync.WaitGroup
	ch  chan struct{}
	err error
}

func (w *testWriter) Write(ctx context.Context, vl *api.ValueList) error {
	w.got = vl
	w.wg.Done()

	select {
	case <-w.ch:
		return w.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

type testError struct{}

func (testError) Error() string {
	return "test error"
}

func TestFanout(t *testing.T) {
	cases := []struct {
		title         string
		returnError   bool
		cancelContext bool
	}{
		{
			title: "success",
		},
		{
			title:       "error",
			returnError: true,
		},
		{
			title:         "context canceled",
			cancelContext: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var (
				done = make(chan struct{})
				wg   sync.WaitGroup
			)

			var writerError error
			if tc.returnError {
				writerError = testError{}
			}
			writers := []*testWriter{
				{
					wg:  &wg,
					ch:  done,
					err: writerError,
				},
				{
					wg:  &wg,
					ch:  done,
					err: writerError,
				},
			}
			wg.Add(len(writers))

			go func() {
				// wait for all writers to be called, then signal them to return
				wg.Wait()

				if tc.cancelContext {
					cancel()
				} else {
					close(done)
				}
			}()

			want := &api.ValueList{
				Identifier: api.Identifier{
					Host:   "example.com",
					Plugin: "TestFanout",
					Type:   "gauge",
				},
				Values:  []api.Value{api.Gauge(42)},
				DSNames: []string{"value"},
			}

			var f api.Fanout
			for _, w := range writers {
				f = append(f, w)
			}

			err := f.Write(ctx, want)
			switch {
			case tc.returnError && !errors.Is(err, testError{}):
				t.Errorf("Fanout.Write() = %v, want %T", err, testError{})
			case tc.cancelContext && !errors.Is(err, context.Canceled):
				t.Errorf("Fanout.Write() = %v, want %T", err, context.Canceled)
			case !tc.returnError && !tc.cancelContext && err != nil:
				t.Errorf("Fanout.Write() = %v", err)
			}

			for i, w := range writers {
				if want == w.got {
					t.Errorf("writers[%d].vl == w.got, want copy", i)
				}
				if diff := cmp.Diff(want, w.got); diff != "" {
					t.Errorf("writers[%d].vl differs (+got/-want):\n%s", i, diff)
				}
			}
		})
	}
}
