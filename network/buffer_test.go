package network

import (
	"math"
	"reflect"
	"testing"

	"collectd.org/api"
)

func TestWriteValues(t *testing.T) {
	b := NewBuffer()

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
	b := NewBuffer()

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
	b := NewBuffer()

	b.writeInt(23, int64(384))

	want := []byte{0, 23, // pkg type
		0, 12, // pkg len
		0, 0, 0, 0, 0, 0, 1, 128, // 384
	}
	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
