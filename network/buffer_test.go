package network

import (
	"reflect"
	"testing"
)

func TestWriteInt(t *testing.T) {
	b := NewBuffer()

	b.writeInt(23, int64(384))

	want := make([]byte, 12, 12)
	want[1] = byte(23) // type
	want[3] = byte(12) // length
	want[10] = byte(1)
	want[11] = byte(128)

	got := b.buffer.Bytes()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
