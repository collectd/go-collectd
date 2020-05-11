package api // import "collectd.org/api"

import (
	"fmt"
	"testing"
)

func TestParseIdentifier(t *testing.T) {
	cases := []struct {
		Input string
		Want  Identifier
	}{
		{
			Input: "example.com/golang/gauge",
			Want: Identifier{
				Host:   "example.com",
				Plugin: "golang",
				Type:   "gauge",
			},
		},
		{
			Input: "example.com/golang-foo/gauge-bar",
			Want: Identifier{
				Host:           "example.com",
				Plugin:         "golang",
				PluginInstance: "foo",
				Type:           "gauge",
				TypeInstance:   "bar",
			},
		},
		{
			Input: "example.com/golang-a-b/gauge-b-c",
			Want: Identifier{
				Host:           "example.com",
				Plugin:         "golang",
				PluginInstance: "a-b",
				Type:           "gauge",
				TypeInstance:   "b-c",
			},
		},
	}

	for i, c := range cases {
		if got, err := ParseIdentifier(c.Input); got != c.Want || err != nil {
			t.Errorf("case %d: got (%v, %v), want (%v, %v)", i, got, err, c.Want, nil)
		}
	}

	failures := []string{
		"example.com/golang",
		"example.com/golang/gauge/extra",
	}

	for _, c := range failures {
		if got, err := ParseIdentifier(c); err == nil {
			t.Errorf("got (%v, %v), want (%v, !%v)", got, err, Identifier{}, nil)
		}
	}
}

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

func TestConfig_Unmarshal(t *testing.T) {
	type dstConf2 struct {
		Args      string
		KeepAlive bool
		Expect    []string
	}
	type dstConf struct {
		Args string
		Host dstConf2
	}
	tests := []struct {
		name    string
		src     Config
		dst     interface{}
		wantErr bool
		verify  func(*dstConf) error
	}{
		{
			name: "Base test",
			src: Config{
				Key: "myPlugin",
				Children: []Config{
					{
						Key: "Host",
						Children: []Config{
							{
								Key: "KeepAlive",
								Values: []ConfigValue{
									{
										typ: configTypeBoolean,
										b:   true,
									},
								},
							},
							{
								Key: "Expect",
								Values: []ConfigValue{
									{
										typ: configTypeString,
										s:   "foo",
									},
								},
							},
							{
								Key: "Expect",
								Values: []ConfigValue{
									{
										typ: configTypeString,
										s:   "bar",
									},
								},
							},
						},
					},
				},
			},
			dst: &dstConf{
				Host: dstConf2{},
			},
			wantErr: false,
			verify: func(d *dstConf) error {
				var ok bool
				for _, v := range d.Host.Expect {
					if v == "foo" {
						ok = true
					}
				}
				if !ok {
					return fmt.Errorf("Host.Expect didn't contain expected value")
				}
				for _, v := range d.Host.Expect {
					if v == "bar" {
						return nil
					}
				}
				return fmt.Errorf("Host.Expect didn't contain expected value")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Key:      tt.src.Key,
				Values:   tt.src.Values,
				Children: tt.src.Children,
			}
			if err := c.Unmarshal(tt.dst); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.verify(tt.dst.(*dstConf)); err != nil {
				t.Errorf("Unmarshal() verify error = %v", err)
			}
		})
	}
}
