package plugin_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/meta"
	"collectd.org/plugin"
	"collectd.org/plugin/fake"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestInterval(t *testing.T) {
	fake.SetInterval(42 * time.Second)
	defer fake.TearDown()

	got, err := plugin.Interval()
	if err != nil {
		t.Fatal(err)
	}

	if want := 42 * time.Second; got != want {
		t.Errorf("Interval() = %v, want %v", got, want)
	}
}

type testLogger struct {
	Name string
	plugin.Severity
	Message string
}

func (l *testLogger) Log(ctx context.Context, s plugin.Severity, msg string) {
	l.Severity = s
	l.Message = msg
	l.Name, _ = plugin.Name(ctx)
}

func TestLog(t *testing.T) {
	cases := []struct {
		title    string
		logFunc  func(v ...interface{}) error
		fmtFunc  func(format string, v ...interface{}) error
		severity plugin.Severity
	}{
		{"Error", plugin.Error, plugin.Errorf, plugin.SeverityError},
		{"Warning", plugin.Warning, plugin.Warningf, plugin.SeverityWarning},
		{"Notice", plugin.Notice, plugin.Noticef, plugin.SeverityNotice},
		{"Info", plugin.Info, plugin.Infof, plugin.SeverityInfo},
		{"Debug", plugin.Debug, plugin.Debugf, plugin.SeverityDebug},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			defer fake.TearDown()

			name := "TestLog_" + tc.title
			l := &testLogger{}
			if err := plugin.RegisterLog(name, l); err != nil {
				t.Fatal(err)
			}

			tc.logFunc("test %d %%s", 42)
			if got, want := l.Name, name; got != want {
				t.Errorf("plugin.Name() = %q, want %q", got, want)
			}
			if got, want := l.Severity, tc.severity; got != want {
				t.Errorf("Severity = %v, want %v", got, want)
			}
			if got, want := l.Message, "test %d %%s42"; got != want {
				t.Errorf("Message = %q, want %q", got, want)
			}

			*l = testLogger{}
			tc.fmtFunc("test %d %%s", 42)
			if got, want := l.Name, name; got != want {
				t.Errorf("plugin.Name() = %q, want %q", got, want)
			}
			if got, want := l.Severity, tc.severity; got != want {
				t.Errorf("Severity = %v, want %v", got, want)
			}
			if got, want := l.Message, "test 42 %s"; got != want {
				t.Errorf("Message = %q, want %q", got, want)
			}
		})
	}
}

func TestRegisterRead(t *testing.T) {
	cases := []struct {
		title        string
		opts         []plugin.ReadOption
		wantGroup    string
		wantInterval time.Duration
	}{
		{
			title:        "default case",
			wantGroup:    "golang",
			wantInterval: 10 * time.Second,
		},
		{
			title:        "with interval",
			opts:         []plugin.ReadOption{plugin.WithInterval(20 * time.Second)},
			wantGroup:    "golang",
			wantInterval: 20 * time.Second,
		},
		{
			title:        "with group",
			opts:         []plugin.ReadOption{plugin.WithGroup("testing")},
			wantGroup:    "testing",
			wantInterval: 10 * time.Second,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			defer fake.TearDown()

			if err := plugin.RegisterRead("TestRegisterRead", &testReader{}, tc.opts...); err != nil {
				t.Fatal(err)
			}

			callbacks := fake.ReadCallbacks()
			if got, want := len(callbacks), 1; got != want {
				t.Errorf("len(ReadCallbacks) = %d, want %d", got, want)
			}
			if len(callbacks) < 1 {
				t.FailNow()
			}

			cb := callbacks[0]
			if got, want := cb.Group, tc.wantGroup; got != want {
				t.Errorf("ReadCallback.Group = %q, want %q", got, want)
			}
			if got, want := cb.Interval.Duration(), tc.wantInterval; got != want {
				t.Errorf("ReadCallback.Interval = %v, want %v", got, want)
			}
		})
	}
}

func TestReadWrite(t *testing.T) {
	baseVL := api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "TestRead",
			Type:   "gauge",
		},
		Time:     time.Unix(1587500000, 0),
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Gauge(42)},
		DSNames:  []string{"value"},
	}

	cases := []struct {
		title    string
		modifyVL func(*api.ValueList)
		readErr  error
		writeErr error
		wantErr  bool
	}{
		{
			title: "gauge",
		},
		{
			title: "gauge NaN",
			modifyVL: func(vl *api.ValueList) {
				vl.Values = []api.Value{api.Gauge(math.NaN())}
			},
		},
		{
			title: "derive",
			modifyVL: func(vl *api.ValueList) {
				vl.Type = "derive"
				vl.Values = []api.Value{api.Derive(42)}
			},
		},
		{
			title: "counter",
			modifyVL: func(vl *api.ValueList) {
				vl.Type = "counter"
				vl.Values = []api.Value{api.Counter(42)}
			},
		},
		{
			title: "bool meta data",
			modifyVL: func(vl *api.ValueList) {
				vl.Meta = meta.Data{
					"key": meta.Bool(true),
				}
			},
		},
		{
			title: "float64 meta data",
			modifyVL: func(vl *api.ValueList) {
				vl.Meta = meta.Data{
					"key": meta.Float64(20.0 / 3.0),
				}
			},
		},
		{
			title: "float64 NaN meta data",
			modifyVL: func(vl *api.ValueList) {
				vl.Meta = meta.Data{
					"key": meta.Float64(math.NaN()),
				}
			},
		},
		{
			title: "int64 meta data",
			modifyVL: func(vl *api.ValueList) {
				vl.Meta = meta.Data{
					"key": meta.Int64(-23),
				}
			},
		},
		{
			title: "uint64 meta data",
			modifyVL: func(vl *api.ValueList) {
				vl.Meta = meta.Data{
					"key": meta.UInt64(42),
				}
			},
		},
		{
			title: "string meta data",
			modifyVL: func(vl *api.ValueList) {
				vl.Meta = meta.Data{
					"key": meta.String(`\\\ value ///`),
				}
			},
		},
		{
			title: "marshaling error",
			modifyVL: func(vl *api.ValueList) {
				vl.Values = []api.Value{nil}
			},
			wantErr: true,
		},
		{
			title: "read callback sets errno",
			// The "plugin_dispatch_values()" implementation of the "fake" package only supports the types
			// "derive", "gauge", and "counter". If another type is encountered, errno is set to EINVAL.
			modifyVL: func(vl *api.ValueList) {
				vl.Type = "invalid"
			},
			wantErr: true,
		},
		{
			title:   "read callback returns error",
			readErr: errors.New("read error"),
			wantErr: true,
		},
		{
			title: "read callback canceled context",
			// Calling plugin.Write() with a canceled context results in an error.
			readErr: context.Canceled,
			wantErr: true,
		},
		{
			title:    "write callback returns error",
			writeErr: errors.New("write error"),
			wantErr:  true,
		},
		{
			title: "plugin name is filled in",
			modifyVL: func(vl *api.ValueList) {
				vl.Plugin = ""
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			defer fake.TearDown()

			vl := baseVL.Clone()
			if tc.modifyVL != nil {
				tc.modifyVL(vl)
			}

			r := &testReader{
				vl:       vl,
				wantName: "TestRead",
				wantErr:  tc.readErr,
			}
			if err := plugin.RegisterRead("TestRead", r); err != nil {
				t.Fatal(err)
			}

			w := &testWriter{
				wantName: "TestWrite",
				wantErr:  tc.writeErr,
			}
			if err := plugin.RegisterWrite("TestWrite", w); err != nil {
				t.Fatal(err)
			}

			err := fake.ReadAll()
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("ReadAll() = %v, want error: %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			if got, want := len(w.valueLists), 1; got != want {
				t.Errorf("len(testWriter.valueLists) = %d, want %d", got, want)
			}
			if len(w.valueLists) < 1 {
				t.FailNow()
			}

			// Expect vl.Plugin to get populated.
			if vl.Plugin == "" {
				vl.Plugin = "TestRead"
			}

			opts := []cmp.Option{
				// cmp complains about meta.Entry having private fields.
				cmp.Transformer("meta.Entry", func(e meta.Entry) interface{} {
					return e.Interface()
				}),
				// transform api.Gauge to float64, so EquateNaNs applies to them.
				cmp.Transformer("api.Gauge", func(g api.Gauge) interface{} {
					return float64(g)
				}),
				cmpopts.EquateNaNs(),
			}
			if got, want := w.valueLists[0], vl; !cmp.Equal(got, want, opts...) {
				t.Errorf("ValueList differs (-want/+got): %s", cmp.Diff(want, got, opts...))
			}
		})
	}
}

type testReader struct {
	vl       *api.ValueList
	wantName string
	wantErr  error
}

func (r *testReader) Read(ctx context.Context) error {
	// Verify that plugin.Name() works inside Read callbacks.
	gotName, ok := plugin.Name(ctx)
	if !ok || gotName != r.wantName {
		return fmt.Errorf("plugin.Name() = (%q, %v), want (%q, %v)", gotName, ok, r.wantName, true)
	}

	if errors.Is(r.wantErr, context.Canceled) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		cancel()
		// continue with canceled context
	} else if r.wantErr != nil {
		return r.wantErr
	}

	return plugin.Write(ctx, r.vl)
}

type testWriter struct {
	valueLists []*api.ValueList
	wantName   string
	wantErr    error
}

func (w *testWriter) Write(ctx context.Context, vl *api.ValueList) error {
	// Verify that plugin.Name() works inside Write callbacks.
	gotName, ok := plugin.Name(ctx)
	if !ok || gotName != w.wantName {
		return fmt.Errorf("plugin.Name() = (%q, %v), want (%q, %v)", gotName, ok, w.wantName, true)
	}

	if w.wantErr != nil {
		return w.wantErr
	}

	w.valueLists = append(w.valueLists, vl)
	return nil
}

func TestShutdown(t *testing.T) {
	// NOTE: fake.TearDown() will remove all callbacks from the C code's state. plugin.shutdownFuncs will still hold
	// a reference to the registered shutdown calls, preventing it from registering another C callback in later
	// tests. Long story short, don't use shutdown callbacks in any other test.
	defer fake.TearDown()

	shutters := []*testShutter{}
	// This creates 20 shutdown functions: one will succeed, 19 will fail.
	// We expect *all* shutdown functions to be called.
	for i := 0; i < 20; i++ {
		s := &testShutter{
			wantName: "TestShutdown",
		}
		callbackName := "TestShutdown"
		if i != 0 {
			callbackName = fmt.Sprintf("failing_function_%d", i)
		}

		if err := plugin.RegisterShutdown(callbackName, s); err != nil {
			t.Fatal(err)
		}
	}

	if err := fake.ShutdownAll(); err == nil {
		t.Error("fake.ShutdownAll() succeeded, expected it to fail")
	}

	for _, s := range shutters {
		if got, want := s.callCount, 1; got != want {
			t.Errorf("testShutter.callCount = %d, want %d", got, want)
		}
	}
}

type testShutter struct {
	wantName  string
	callCount int
}

func (s *testShutter) Shutdown(ctx context.Context) error {
	s.callCount++

	// Verify that plugin.Name() works inside Shutdown callbacks.
	gotName, ok := plugin.Name(ctx)
	if !ok || gotName != s.wantName {
		return fmt.Errorf("plugin.Name() = (%q, %v), want (%q, %v)", gotName, ok, s.wantName, true)
	}

	return nil
}
