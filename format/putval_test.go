package format_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"collectd.org/api"
	"collectd.org/format"
)

func TestPutval(t *testing.T) {
	baseVL := api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "TestPutval",
			Type:   "derive",
		},
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Derive(42)},
		DSNames:  []string{"value"},
	}

	cases := []struct {
		title   string
		modify  func(*api.ValueList)
		want    string
		wantErr bool
	}{
		{
			title: "derive",
			want:  `PUTVAL "example.com/TestPutval/derive" interval=10.000 N:42` + "\n",
		},
		{
			title: "gauge",
			modify: func(vl *api.ValueList) {
				vl.Type = "gauge"
				vl.Values = []api.Value{api.Gauge(20.0 / 3.0)}
			},
			want: `PUTVAL "example.com/TestPutval/gauge" interval=10.000 N:6.66666666666667` + "\n",
		},
		{
			title: "counter",
			modify: func(vl *api.ValueList) {
				vl.Type = "counter"
				vl.Values = []api.Value{api.Counter(31337)}
			},
			want: `PUTVAL "example.com/TestPutval/counter" interval=10.000 N:31337` + "\n",
		},
		{
			title: "multiple values",
			modify: func(vl *api.ValueList) {
				vl.Type = "if_octets"
				vl.Values = []api.Value{api.Derive(1), api.Derive(2)}
				vl.DSNames = []string{"rx", "tx"}
			},
			want: `PUTVAL "example.com/TestPutval/if_octets" interval=10.000 N:1:2` + "\n",
		},
		{
			title: "invalid type",
			modify: func(vl *api.ValueList) {
				vl.Values = []api.Value{nil}
			},
			wantErr: true,
		},
		{
			title: "time",
			modify: func(vl *api.ValueList) {
				vl.Time = time.Unix(1588087972, 987654321)
			},
			want: `PUTVAL "example.com/TestPutval/derive" interval=10.000 1588087972.988:42` + "\n",
		},
		{
			title: "interval",
			modify: func(vl *api.ValueList) {
				vl.Interval = 9876543 * time.Microsecond
			},
			want: `PUTVAL "example.com/TestPutval/derive" interval=9.877 N:42` + "\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			ctx := context.Background()

			vl := baseVL
			if tc.modify != nil {
				tc.modify(&vl)
			}

			var b strings.Builder
			err := format.NewPutval(&b).Write(ctx, &vl)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("Putval.Write(%#v) = %v, want error %v", &vl, err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			if got := b.String(); got != tc.want {
				t.Errorf("Putval.Write(%#v): got %q, want %q", &vl, got, tc.want)
			}
		})
	}
}
