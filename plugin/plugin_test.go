package plugin_test

import (
	"context"
	"testing"

	"collectd.org/plugin"
	"collectd.org/plugin/fake"
)

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
