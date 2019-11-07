// Package format provides utilities to format metrics and notifications in
// various formats.
package format // import "collectd.org/format"

// Author: Remi Ferrand <remi.ferrand_at_cc.in2p3.fr>

import (
	"context"
	"fmt"
	"io"
	"strings"

	"collectd.org/api"
)

// PutvalWithMeta implements the Writer interface for PutvalWithMeta formatted output.
// This format is not a standard format like PUTVAL is.
// This formatter is currently only intended to help while developing plugins that support
// metadata (LISTVAL or GETVAL does not currently display them)
type PutvalWithMeta struct {
	w io.Writer
}

// NewPutvalWithMeta returns a new PutvalWithMeta object writing to the provided io.Writer.
func NewPutvalWithMeta(w io.Writer) *PutvalWithMeta {
	return &PutvalWithMeta{
		w: w,
	}
}

// Write formats the ValueList in the PutvalWithMeta format and writes it to the
// assiciated io.Writer.
func (p *PutvalWithMeta) Write(_ context.Context, vl *api.ValueList) error {
	s, err := formatValues(vl)
	if err != nil {
		return err
	}

	var metaStr string
	if len(vl.Metadata) > 0 {
		metaKeys := vl.Metadata.Toc()
		metaPairs := make([]string, len(vl.Metadata))
		for i, key := range metaKeys {
			metaPairs[i] = fmt.Sprintf("%s=\"%s\"", key, vl.Metadata.GetAsString(key))
		}

		metaStr = " {" + strings.Join(metaPairs, ",") + "}"
	}

	_, err = fmt.Fprintf(p.w, "PUTVAL %q interval=%.3f %s%s\n",
		vl.Identifier.String(), vl.Interval.Seconds(), s, metaStr)
	return err
}
