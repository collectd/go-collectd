package api

import (
	"fmt"
	"testing"
)

type dstConf2 struct {
	Args      string
	KeepAlive bool
	Expect    []string
	Hats      int8
}
type dstConf struct {
	Args string
	Host dstConf2
}
type dstConf3 struct {
	Args string
	Host []dstConf2
}

func TestConfig_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		src     Config
		dst     interface{}
		wantErr bool
		verify  func(interface{}) error
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
							{
								Key: "Hats",
								Values: []ConfigValue{
									{
										typ: configTypeNumber,
										f:   424242.42,
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
			verify: func(i interface{}) error {
				d := i.(*dstConf)
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
		{
			name: "Test slice of struct",
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
							{
								Key: "Hats",
								Values: []ConfigValue{
									{
										typ: configTypeNumber,
										f:   424242.42,
									},
								},
							},
						},
					},
				},
			},
			dst:     &dstConf3{},
			wantErr: false,
			verify: func(i interface{}) error {
				d := i.(*dstConf3)
				var ok bool
				for _, v := range d.Host[0].Expect {
					if v == "foo" {
						ok = true
					}
				}
				if !ok {
					return fmt.Errorf("Host.Expect didn't contain expected value")
				}
				for _, v := range d.Host[0].Expect {
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
			switch dst := tt.dst.(type) {
			case *dstConf:
				if err := tt.verify(dst); err != nil {
					t.Errorf("Unmarshal() verify error = %v", err)
				}
			case *dstConf3:
				if err := tt.verify(tt.dst.(*dstConf3)); err != nil {
					t.Errorf("Unmarshal() verify error = %v", err)
				}
			}
		})
	}
}
