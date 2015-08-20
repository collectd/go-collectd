// +build gofuzz

package network // import "collectd.org/network"

import (
	"bytes"
	"fmt"
)

// Fuzz is used by the https://github.com/dvyukov/go-fuzz framework
// It's method signature must match the prescribed format and it is expected to panic upon failure
// Usage:
//   $ go-fuzz-build collectd.org/network
//   $ mkdir -p /tmp/fuzzwork/corpus
//   $ cp network/testdata/packet1.bin /tmp/fuzzwork/corpus
//   $ go-fuzz -bin=./network-fuzz.zip -workdir=/tmp/fuzzwork
func Fuzz(data []byte) int {

	// deserialize
	d1, err := Parse(data, ParseOpts{})
	if err != nil {
		return 0
	}
	if len(d1) == 0 && err != nil {
		panic("d1 is empty but no err was returned")
	}

	// serialize
	s1 := NewBuffer(0)
	if err := s1.Write(d1[0]); err != nil {
		panic(err)
	}

	// deserialize
	d2, err := Parse(s1.buffer.Bytes(), ParseOpts{})
	if err != nil {
		return 0
	}
	if len(d2) == 0 && err != nil {
		panic("d2 is empty but no err was returned")
	}

	// serialize
	s2 := NewBuffer(0)
	if err := s2.Write(d2[0]); err != nil {
		panic(err)
	}

	if bytes.Compare(s1.buffer.Bytes(), s2.buffer.Bytes()) != 0 {
		panic(fmt.Sprintf("Comparison of two serialized versions failed s1 [%v] s2[%v]", s1.buffer.Bytes(), s2.buffer.Bytes()))
	}

	return 1
}
