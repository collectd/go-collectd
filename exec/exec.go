// Package exec implements tools to write plugins for collectd's "exec plugin"
// in Go.
package exec // import "collectd.org/exec"

import (
	"context"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"collectd.org/api"
	"collectd.org/format"
)

// Putval is the dispatcher used by the exec package to print ValueLists.
var Putval api.Writer = format.NewPutval(os.Stdout)

type callback interface {
	run(context.Context, *sync.WaitGroup)
	stop()
}

// Executor holds one or more callbacks which are called periodically.
type Executor struct {
	cb    []callback
	group sync.WaitGroup
}

// NewExecutor returns a pointer to a new Executor object.
func NewExecutor() *Executor {
	return &Executor{
		group: sync.WaitGroup{},
	}
}

// ValueCallback adds a simple "value" callback to the Executor. The callback
// only returns a Number, i.e. either a api.Gauge or api.Derive, and formatting
// and printing is done by the executor.
func (e *Executor) ValueCallback(callback func() api.Value, vl *api.ValueList) {
	e.cb = append(e.cb, &valueCallback{
		callback: callback,
		vl:       *vl,
		done:     make(chan struct{}),
	})
}

// VoidCallback adds a "complex" callback to the Executor. While the functions
// prototype is simpler, all the work has to be done by the callback, i.e. the
// callback needs to format and print the appropriate lines to "STDOUT".
// However, this allows cases in which the number of values reported varies,
// e.g. depending on the system the code is running on.
func (e *Executor) VoidCallback(callback func(context.Context, time.Duration), interval time.Duration) {
	e.cb = append(e.cb, voidCallback{
		callback: callback,
		interval: interval,
		done:     make(chan struct{}),
	})
}

// Run starts calling all callbacks periodically and blocks.
func (e *Executor) Run(ctx context.Context) {
	for _, cb := range e.cb {
		e.group.Add(1)
		go cb.run(ctx, &e.group)
	}

	e.group.Wait()
}

// Stop sends a signal to all callbacks to exit and returns. This unblocks
// "Run()" but does not block itself.
func (e *Executor) Stop() {
	for _, cb := range e.cb {
		cb.stop()
	}
}

type valueCallback struct {
	callback func() api.Value
	vl       api.ValueList
	done     chan struct{}
}

func (cb *valueCallback) run(ctx context.Context, g *sync.WaitGroup) {
	defer g.Done()

	if cb.vl.Host == "" {
		cb.vl.Host = Hostname()
	}
	cb.vl.Interval = sanitizeInterval(cb.vl.Interval)

	ticker := time.NewTicker(cb.vl.Interval)
	for {
		select {
		case <-ticker.C:
			cb.vl.Values = []api.Value{cb.callback()}
			cb.vl.Time = time.Now()
			Putval.Write(ctx, &cb.vl)
		case <-cb.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (cb *valueCallback) stop() {
	close(cb.done)
}

type voidCallback struct {
	callback func(context.Context, time.Duration)
	interval time.Duration
	done     chan struct{}
}

func (cb voidCallback) run(ctx context.Context, g *sync.WaitGroup) {
	defer g.Done()

	ticker := time.NewTicker(sanitizeInterval(cb.interval))

	for {
		select {
		case <-ticker.C:
			cb.callback(ctx, cb.interval)
		case <-cb.done:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (cb voidCallback) stop() {
	close(cb.done)
}

// Interval determines the default interval from the "COLLECTD_INTERVAL"
// environment variable. It falls back to 10s if the environment variable is
// unset or cannot be parsed.
func Interval() time.Duration {
	i, err := strconv.ParseFloat(os.Getenv("COLLECTD_INTERVAL"), 64)
	if err != nil {
		log.Printf("unable to determine default interval: %v", err)
		return time.Second * 10
	}

	return time.Duration(i * float64(time.Second))
}

// Hostname determines the hostname to use from the "COLLECTD_HOSTNAME"
// environment variable and falls back to os.Hostname() if it is unset. If that
// also fails an empty string is returned.
func Hostname() string {
	if h := os.Getenv("COLLECTD_HOSTNAME"); h != "" {
		return h
	}

	if h, err := os.Hostname(); err == nil {
		return h
	}

	return ""
}

func sanitizeInterval(in time.Duration) time.Duration {
	if in == time.Duration(0) {
		return Interval()
	}

	return in
}
