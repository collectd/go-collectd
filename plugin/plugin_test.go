package plugin_test

import (
	"context"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/plugin"
	"collectd.org/plugin/fake"
	"github.com/google/go-cmp/cmp"
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
	defer fake.TearDown()

	l := &testLogger{}
	if err := plugin.RegisterLog("TestLog", l); err != nil {
		t.Fatal(err)
	}

	plugin.Infof("test %d %%s", 42)

	if got, want := l.Severity, plugin.SeverityInfo; got != want {
		t.Errorf("Severity = %v, want %v", got, want)
	}

	if got, want := l.Message, "test 42 %s"; got != want {
		t.Errorf("Message = %q, want %q", got, want)
	}

	if got, want := l.Name, "TestLog"; got != want {
		t.Errorf("plugin.Name() = %q, want %q", got, want)
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
		wantErr  bool
	}{
		{
			title: "gauge",
		},
		{
			title: "derive",
			modifyVL: func(vl *api.ValueList) {
				vl.Type = "derive"
				vl.Values[0] = api.Derive(42)
			},
		},
		{
			title: "counter",
			modifyVL: func(vl *api.ValueList) {
				vl.Type = "counter"
				vl.Values[0] = api.Counter(42)
			},
		},
		{
			title: "invalid type",
			modifyVL: func(vl *api.ValueList) {
				vl.Type = "invalid"
			},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			defer fake.TearDown()

			vl := baseVL
			if tc.modifyVL != nil {
				tc.modifyVL(&vl)
			}

			r := &testReader{
				vl: &vl,
			}
			if err := plugin.RegisterRead("TestRead", r); err != nil {
				t.Fatal(err)
			}

			w := &testWriter{}
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

			if got, want := w.valueLists[0], &vl; !cmp.Equal(got, want) {
				t.Errorf("ValueList differs (-want/+got): %s", cmp.Diff(want, got))
			}
		})
	}
}

type testReader struct {
	vl *api.ValueList
}

func (r *testReader) Read(ctx context.Context) error {
	w := plugin.Writer{}
	return w.Write(ctx, r.vl)
}

type testWriter struct {
	valueLists []*api.ValueList
}

func (w *testWriter) Write(ctx context.Context, vl *api.ValueList) error {
	w.valueLists = append(w.valueLists, vl)
	return nil
}
