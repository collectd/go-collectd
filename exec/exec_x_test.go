package exec_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/exec"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type testWriter struct {
	vl *api.ValueList
}

func (w *testWriter) Write(_ context.Context, vl *api.ValueList) error {
	if w.vl != nil {
		return errors.New("received unexpected second value")
	}

	w.vl = vl
	return nil
}

func TestValueCallback_ExecutorStop(t *testing.T) {
	cases := []struct {
		title    string
		stopFunc func(f context.CancelFunc, e *exec.Executor)
	}{
		{"ExecutorStop", func(_ context.CancelFunc, e *exec.Executor) { e.Stop() }},
		{"CancelContext", func(cancel context.CancelFunc, _ *exec.Executor) { cancel() }},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := os.Setenv("COLLECTD_HOSTNAME", "example.com"); err != nil {
				t.Fatal(err)
			}
			defer func() {
				os.Unsetenv("COLLECTD_HOSTNAME")
			}()

			savedPutval := exec.Putval
			defer func() {
				exec.Putval = savedPutval
			}()

			w := &testWriter{}
			exec.Putval = w

			e := exec.NewExecutor()
			ch := make(chan struct{})
			go func() {
				// wait for ch to be closed
				<-ch
				tc.stopFunc(cancel, e)
			}()

			e.ValueCallback(func() api.Value {
				defer func() {
					close(ch)
				}()
				return api.Derive(42)
			}, &api.ValueList{
				Identifier: api.Identifier{
					Plugin: "go-exec",
					Type:   "derive",
				},
				Interval: time.Millisecond,
				DSNames:  []string{"value"},
			})

			// e.Run() blocks until the context is canceled or
			// e.Stop() is called (see tc.stopFunc above).
			e.Run(ctx)

			want := &api.ValueList{
				Identifier: api.Identifier{
					Host:   "example.com",
					Plugin: "go-exec",
					Type:   "derive",
				},
				Interval: time.Millisecond,
				Values:   []api.Value{api.Derive(42)},
				DSNames:  []string{"value"},
			}
			if diff := cmp.Diff(want, w.vl, cmpopts.IgnoreFields(api.ValueList{}, "Time")); diff != "" {
				t.Errorf("received value lists differ (+got/-want):\n%s", diff)
			}
		})
	}
}

func TestVoidCallback(t *testing.T) {
	cases := []struct {
		title    string
		stopFunc func(f context.CancelFunc, e *exec.Executor)
	}{
		{"ExecutorStop", func(_ context.CancelFunc, e *exec.Executor) { e.Stop() }},
		{"CancelContext", func(cancel context.CancelFunc, _ *exec.Executor) { cancel() }},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			e := exec.NewExecutor()
			ch := make(chan struct{})
			go func() {
				// wait for ch to be closed
				<-ch
				tc.stopFunc(cancel, e)
			}()

			var calls int
			e.VoidCallback(func(_ context.Context, d time.Duration) {
				if got, want := d, time.Millisecond; got != want {
					t.Errorf("VoidCallback(%v), want argument %v", got, want)
				}
				calls++

				close(ch)
			}, time.Millisecond)

			// e.Run() blocks until the context is canceled or
			// e.Stop() is called (see tc.stopFunc above).
			e.Run(ctx)

			if got, want := calls, 1; got != want {
				t.Errorf("number of calls = %d, want %d", got, want)
			}
		})
	}
}
