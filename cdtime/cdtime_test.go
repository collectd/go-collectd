package cdtime // import "collectd.org/cdtime"

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
	"time"
)

func TestDecode(t *testing.T) {

	buf, err := hex.DecodeString("156DB1CFAF08BA7A0009000C")
	if err != nil {
		panic(err)
	}

	var i uint64
	b := bytes.NewBuffer(buf)
	if err := binary.Read(b, binary.BigEndian, &i); err != nil {
		panic(err)
	}

	hrtime := Time(i)

	want := "2015-07-27 17:05:18.73490774 -0700 PDT"
	if hrtime.Time().String() != want {
		t.Errorf("got %v, want %v", hrtime.Time(), want)
	}
}

func TestSerDes(t *testing.T) {
	for _, date := range []string{
		"2009-02-04T21:00:57-08:00",
		"2009-02-04T21:00:57.1-08:00",
		"2009-02-04T21:00:57.01-08:00",
		"2009-02-04T21:00:57.001-08:00",
		"2009-02-04T21:00:57.0001-08:00",
		"2009-02-04T21:00:57.00001-08:00",
		"2009-02-04T21:00:57.000001-08:00",
		"2009-02-04T21:00:57.0000001-08:00",
		"2009-02-04T21:00:57.00000001-08:00",
		"2009-02-04T21:00:57.000000001-08:00"} {
		testSerDes(t, date)
	}
}

func testSerDes(t *testing.T, date string) {
	t1, err := time.Parse(time.RFC3339Nano, date)
	if err != nil {
		panic(err)
	}

	t2 := New(t1)
	if t1.UnixNano() != t2.Time().UnixNano() {
		t.Errorf("got %s, want %s\n", t2.Time().Format(time.RFC3339Nano), t1.Format(time.RFC3339Nano))
	}
}

func TestNew(t *testing.T) {

	want := time.Unix(1438062739, 123456788)
	got := New(want).Time()

	if want != got {
		t.Errorf("got %v, want %v", got, want)
	}
}
