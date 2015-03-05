// Package format provides utilities to format metrics and notifications in
// various formats.
package format

import (
	"fmt"
	"strings"

	"collectd.org/api"
)

func formatValues(vl api.ValueList) string {
	fields := make([]string, 1+len(vl.Values))

	fields[0] = "N"
	if !vl.Time.IsZero() {
		fields[0] = fmt.Sprintf("%.3f", float64(vl.Time.UnixNano())/1000000000.0)
	}

	for i, v := range vl.Values {
		switch v.(type) {
		case api.Gauge:
			fields[i+1] = fmt.Sprintf("%g", v)
		case api.Derive:
			fields[i+1] = fmt.Sprintf("%d", v)
		default:
			panic(fmt.Errorf("Number has unexpected type: %#v", v))
		}
	}

	return strings.Join(fields, ":")
}

// Putval formats the ValueList in the "PUTVAL" format.
func Putval(vl api.ValueList) string {
	return fmt.Sprintf("PUTVAL %q interval=%.3f %s\n", vl.Identifier.String(), vl.Interval.Seconds(), formatValues(vl))
}
