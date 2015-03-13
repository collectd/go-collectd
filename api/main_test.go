package api // import "collectd.org/api"

import (
	"testing"
)

func TestIdentifierString(t *testing.T) {
	id := Identifier{
		Host:   "example.com",
		Plugin: "golang",
		Type:   "gauge",
	}

	cases := []struct {
		PluginInstance, TypeInstance string
		Want                         string
	}{
		{"", "", "example.com/golang/gauge"},
		{"foo", "", "example.com/golang-foo/gauge"},
		{"", "foo", "example.com/golang/gauge-foo"},
		{"foo", "bar", "example.com/golang-foo/gauge-bar"},
	}

	for _, c := range cases {
		id.PluginInstance = c.PluginInstance
		id.TypeInstance = c.TypeInstance

		got := id.String()
		if got != c.Want {
			t.Errorf("got %q, want %q", got, c.Want)
		}
	}
}
