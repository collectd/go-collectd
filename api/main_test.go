package api_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

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

func TestValueList_Check(t *testing.T) {
	baseVL := api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "TestValueList_Check",
			Type:   "gauge",
		},
		Time:     time.Unix(1589283551, 0),
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Gauge(42)},
		DSNames:  []string{"value"},
	}

	cases := []struct {
		title   string
		modify  func(vl *api.ValueList)
		wantErr bool
	}{
		{
			title: "success",
		},
		{
			title: "without host",
			modify: func(vl *api.ValueList) {
				vl.Host = ""
			},
			wantErr: true,
		},
		{
			title: "host contains hyphen",
			modify: func(vl *api.ValueList) {
				vl.Host = "example-host.com"
			},
		},
		{
			title: "without plugin",
			modify: func(vl *api.ValueList) {
				vl.Plugin = ""
			},
			wantErr: true,
		},
		{
			title: "plugin contains hyphen",
			modify: func(vl *api.ValueList) {
				vl.Plugin = "TestValueList-Check"
			},
			wantErr: true,
		},
		{
			title: "without type",
			modify: func(vl *api.ValueList) {
				vl.Type = ""
			},
			wantErr: true,
		},
		{
			title: "type contains hyphen",
			modify: func(vl *api.ValueList) {
				vl.Type = "http-request"
			},
			wantErr: true,
		},
		{
			title: "without time",
			modify: func(vl *api.ValueList) {
				vl.Time = time.Time{}
			},
		},
		{
			title: "without interval",
			modify: func(vl *api.ValueList) {
				vl.Interval = 0
			},
			wantErr: true,
		},
		{
			title: "without values",
			modify: func(vl *api.ValueList) {
				vl.Values = nil
			},
			wantErr: true,
		},
		{
			title: "surplus values",
			modify: func(vl *api.ValueList) {
				vl.Values = []api.Value{api.Gauge(1), api.Gauge(2)}
			},
			wantErr: true,
		},
		{
			title: "without dsnames",
			modify: func(vl *api.ValueList) {
				vl.DSNames = nil
			},
		},
		{
			title: "surplus dsnames",
			modify: func(vl *api.ValueList) {
				vl.DSNames = []string{"rx", "tx"}
			},
			wantErr: true,
		},
		{
			title: "multiple values",
			modify: func(vl *api.ValueList) {
				vl.Type = "if_octets"
				vl.Values = []api.Value{api.Derive(0), api.Derive(0)}
				vl.DSNames = []string{"rx", "tx"}
			},
		},
		{
			title: "ds name not unique",
			modify: func(vl *api.ValueList) {
				vl.Type = "if_octets"
				vl.Values = []api.Value{api.Derive(0), api.Derive(0)}
				vl.DSNames = []string{"value", "value"}
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			vl := baseVL.Clone()
			if tc.modify != nil {
				tc.modify(vl)
			}

			err := vl.Check()
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("%#v.Check() = %v, want error %v", vl, err, tc.wantErr)

			}
		})
	}
}
