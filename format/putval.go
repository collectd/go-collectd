// Package format provides utilities to format metrics and notifications in
// various formats.
package format // import "collectd.org/format"

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"collectd.org/api"
	"collectd.org/meta"
)

// Putval implements the Writer interface for PUTVAL formatted output.
type Putval struct {
	w io.Writer
}

// NewPutval returns a new Putval object writing to the provided io.Writer.
func NewPutval(w io.Writer) *Putval {
	return &Putval{
		w: w,
	}
}

// Write formats the ValueList in the PUTVAL format and writes it to the
// assiciated io.Writer.
func (p *Putval) Write(_ context.Context, vl *api.ValueList) error {
	s, err := formatValues(vl)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(p.w, "PUTVAL %q interval=%.3f %s%s\n",
		vl.Identifier.String(), vl.Interval.Seconds(), formatMeta(vl.Meta), s)
	return err
}

func formatValues(vl *api.ValueList) (string, error) {
	fields := make([]string, 1+len(vl.Values))

	fields[0] = formatTime(vl.Time)

	for i, v := range vl.Values {
		switch v := v.(type) {
		case api.Counter:
			fields[i+1] = fmt.Sprintf("%d", v)
		case api.Gauge:
			fields[i+1] = fmt.Sprintf("%.15g", v)
		case api.Derive:
			fields[i+1] = fmt.Sprintf("%d", v)
		default:
			return "", fmt.Errorf("unexpected type %T", v)
		}
	}

	return strings.Join(fields, ":"), nil
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N"
	}

	return fmt.Sprintf("%.3f", float64(t.UnixNano())/1000000000.0)
}

var stringWarning sync.Once

func formatMeta(m meta.Data) string {
	if len(m) == 0 {
		return ""
	}

	var values []string
	for k, v := range m {
		// collectd only supports string meta data values as of 5.11.
		if !v.IsString() {
			stringWarning.Do(func() {
				log.Printf("Non-string metadata not supported yet")
			})
			continue
		}
		values = append(values, fmt.Sprintf("meta:%s=%q ", k, v.String()))
	}

	return strings.Join(values, "")
}
