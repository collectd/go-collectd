package network

import (
	"bytes"
	"math"
	"reflect"
	"testing"
	"time"

	"collectd.org/api"
)

func TestWriteValueList(t *testing.T) {
	gotBuf := new(bytes.Buffer)
	b := NewBuffer(gotBuf)

	vl := api.ValueList{
		Identifier: api.Identifier{
			Host:   "example.com",
			Plugin: "golang",
			Type:   "gauge",
		},
		Time:     time.Unix(1426076671, 123000000), // Wed Mar 11 13:24:31 CET 2015
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Derive(1)},
	}

	if err := b.WriteValueList(vl); err != nil {
		t.Errorf("WriteValueList got %v, want nil", err)
		return
	}

	// ValueList with much the same fields, to test compression.
	vl = api.ValueList{
		Identifier: api.Identifier{
			Host:           "example.com",
			Plugin:         "golang",
			PluginInstance: "test",
			Type:           "gauge",
		},
		Time:     time.Unix(1426076681, 234000000), // Wed Mar 11 13:24:41 CET 2015
		Interval: 10 * time.Second,
		Values:   []api.Value{api.Derive(2)},
	}

	if err := b.WriteValueList(vl); err != nil {
		t.Errorf("WriteValueList got %v, want nil", err)
		return
	}

	want := []byte{
		// vl1
		0, 0, 0, 16, 'e', 'x', 'a', 'm', 'p', 'l', 'e', '.', 'c', 'o', 'm', 0,
		0, 2, 0, 11, 'g', 'o', 'l', 'a', 'n', 'g', 0,
		0, 4, 0, 10, 'g', 'a', 'u', 'g', 'e', 0,
		0, 8, 0, 12, 0x15, 0x40, 0x0c, 0xff, 0xc7, 0xdf, 0x3b, 0x64,
		0, 9, 0, 12, 0, 0, 0, 0x02, 0x80, 0, 0, 0,
		0, 6, 0, 15, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0, 1,
		// vl2
		0, 3, 0, 9, 't', 'e', 's', 't', 0,
		0, 8, 0, 12, 0x15, 0x40, 0x0d, 0x02, 0x4e, 0xf9, 0xdb, 0x22,
		0, 6, 0, 15, 0, 1, 2, 0, 0, 0, 0, 0, 0, 0, 2,
	}
	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWriteTime(t *testing.T) {
	b := &Buffer{buffer: new(bytes.Buffer)}
	b.writeTime(time.Unix(1426083986, 314000000)) // Wed Mar 11 15:26:26 CET 2015

	// 1426083986.314 * 2^30 -> 1531246020641985396.736
	// 1531246020641985396 -> 0x1540142494189374
	want := []byte{0, 8, // pkg type
		0, 12, // pkg len
		0x15, 0x40, 0x14, 0x24, 0x94, 0x18, 0x93, 0x74,
	}
	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWriteValues(t *testing.T) {
	b := &Buffer{buffer: new(bytes.Buffer)}

	b.writeValues([]api.Value{
		api.Gauge(42),
		api.Derive(31337),
		api.Gauge(math.NaN()),
	})

	want := []byte{0, 6, // pkg type
		0, 33, // pkg len
		0, 3, // num values
		1, 2, 1, // gauge, derive, gauge
		0, 0, 0, 0, 0, 0, 0x45, 0x40, // 42.0
		0, 0, 0, 0, 0, 0, 0x7a, 0x69, // 31337
		0, 0, 0, 0, 0, 0, 0xf8, 0x7f, // NaN
	}
	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWriteString(t *testing.T) {
	b := &Buffer{buffer: new(bytes.Buffer)}

	b.writeString(0xf007, "foo")

	want := []byte{0xf0, 0x07, // pkg type
		0, 8, // pkg len
		'f', 'o', 'o', 0, // "foo\0"
	}
	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestWriteInt(t *testing.T) {
	b := &Buffer{buffer: new(bytes.Buffer)}

	b.writeInt(23, uint64(384))

	want := []byte{0, 23, // pkg type
		0, 12, // pkg len
		0, 0, 0, 0, 0, 0, 1, 128, // 384
	}
	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSign(t *testing.T) {
	want := []byte{
		2, 0, 0, 49,
		0xcd, 0xa5, 0x9a, 0x37, 0xb0, 0x81, 0xc2, 0x31,
		0x24, 0x2a, 0x6d, 0xbd, 0xfb, 0x44, 0xdb, 0xd7,
		0x41, 0x2a, 0xf4, 0x29, 0x83, 0xde, 0xa5, 0x11,
		0x96, 0xd2, 0xe9, 0x30, 0x21, 0xae, 0xc5, 0x45,
		'a', 'd', 'm', 'i', 'n',
		'c', 'o', 'l', 'l', 'e', 'c', 't', 'd',
	}
	got := sign([]byte{'c', 'o', 'l', 'l', 'e', 'c', 't', 'd'}, "admin", "admin")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
