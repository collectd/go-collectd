package network

import (
	"reflect"
	"testing"
)

func TestSign(t *testing.T) {
	want := []byte{
		2, 0, 0, 41,
		0xcd, 0xa5, 0x9a, 0x37, 0xb0, 0x81, 0xc2, 0x31,
		0x24, 0x2a, 0x6d, 0xbd, 0xfb, 0x44, 0xdb, 0xd7,
		0x41, 0x2a, 0xf4, 0x29, 0x83, 0xde, 0xa5, 0x11,
		0x96, 0xd2, 0xe9, 0x30, 0x21, 0xae, 0xc5, 0x45,
		'a', 'd', 'm', 'i', 'n',
		'c', 'o', 'l', 'l', 'e', 'c', 't', 'd',
	}
	got := signSHA256([]byte{'c', 'o', 'l', 'l', 'e', 'c', 't', 'd'}, "admin", "admin")

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	userToPassword := map[string]string{
		"admin": "admin",
	}
	ok, err := verifySHA256(want[4:41], want[41:], userToPassword)
	if !ok || err != nil {
		t.Errorf("got (%v, %v), want (true, nil)", ok, err)
	}

	want[41], want[42] = want[42], want[41] // corrupt data
	ok, err = verifySHA256(want[4:41], want[41:], userToPassword)
	if ok || err != nil {
		t.Errorf("got (%v, %v), want (false, nil)", ok, err)
	}

	want[41], want[42] = want[42], want[41] // fix data
	userToPassword["admin"] = "test123"     // different password
	ok, err = verifySHA256(want[4:41], want[41:], userToPassword)
	if ok || err != nil {
		t.Errorf("got (%v, %v), want (false, nil)", ok, err)
	}
}
