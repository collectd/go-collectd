package plugin_test

import (
	"context"
	"testing"
	"time"

	"collectd.org/plugin"
	"collectd.org/plugin/fake"
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
