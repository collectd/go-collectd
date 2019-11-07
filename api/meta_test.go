package api

// Author: Remi Ferrand <remi.ferrand_at_cc.in2p3.fr>

import (
	"reflect"
	"testing"
)

func TestMetadataGetAsString(t *testing.T) {
	t.Parallel()

	meta := make(Metadata)
	meta.Set("int_value", int64(-42))
	meta.Set("bool_value", true)
	meta.Set("str_value", "hello")
	meta.Set("uint_value", uint64(42))
	meta.Set("float_value", float64(42.0))

	testCases := []struct {
		Key      string
		Expected string
	}{
		{"int_value", "-42"},
		{"bool_value", "true"},
		{"str_value", "hello"},
		{"uint_value", "42"},
		{"float_value", "4.20000e+01"},
	}

	for _, testCase := range testCases {
		asString := meta.GetAsString(testCase.Key)
		if asString != testCase.Expected {
			t.Errorf("with key '%s', got '%s' while '%s' was expected",
				testCase.Key, asString, testCase.Expected)
		}
	}
}

func TestMetadataToc(t *testing.T) {
	t.Parallel()

	meta := make(Metadata)
	meta.Set("int_value", int64(-42))
	meta.Set("bool_value", true)

	expected := []string{"bool_value", "int_value"}

	toc := meta.Toc()
	if !reflect.DeepEqual(toc, expected) {
		t.Errorf("%+v != %+v", toc, expected)
	}
}

func TestMetadataCloneMerge(t *testing.T) {
	t.Parallel()

	orig := make(Metadata)
	orig.Set("a", 42)
	orig.Set("b", false)

	additional := make(Metadata)
	additional.Set("c", true)
	additional.Set("b", 54)

	new := orig.CloneMerge(additional)

	expected := make(Metadata)
	expected.Set("c", true)
	expected.Set("b", 54)
	expected.Set("a", 42)

	if !reflect.DeepEqual(new, expected) {
		t.Errorf("%+v != %+v", new, expected)
	}

	t.Run("with empty additional meta", func(t *testing.T) {
		additional := make(Metadata)
		new := orig.CloneMerge(additional)
		if !reflect.DeepEqual(new, orig) {
			t.Errorf("%+v != %+v", new, orig)
		}
	})
}
