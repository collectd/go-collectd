package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
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

func stringPtr(s string) *string {
	return &s
}

func TestConfig_Unmarshal(t *testing.T) {
	tests := []struct {
		name    string
		src     Block
		dst     interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name: "Base test",
			src: Block{
				Key: "myPlugin",
				Children: []Block{
					{
						Key: "Host",
						Children: []Block{
							{
								Key: "KeepAlive",
								Values: []Value{
									{
										typ: booleanType,
										b:   true,
									},
								},
							},
							{
								Key: "Expect",
								Values: []Value{
									{
										typ: stringType,
										s:   "foo",
									},
								},
							},
							{
								Key: "Expect",
								Values: []Value{
									{
										typ: stringType,
										s:   "bar",
									},
								},
							},
							{
								Key: "Hats",
								Values: []Value{
									{
										typ: numberType,
										f:   424242.42,
									},
								},
							},
						},
					},
				},
			},
			dst: &dstConf{},
			want: &dstConf{
				Host: dstConf2{
					KeepAlive: true,
					Expect:    []string{"foo", "bar"},
					Hats:      50, // truncated to 8bit
				},
			},
		},
		{
			name: "Test slice of struct",
			src: Block{
				Key: "myPlugin",
				Children: []Block{
					{
						Key: "Host",
						Children: []Block{
							{
								Key: "KeepAlive",
								Values: []Value{
									{
										typ: booleanType,
										b:   true,
									},
								},
							},
							{
								Key: "Expect",
								Values: []Value{
									{
										typ: stringType,
										s:   "foo",
									},
								},
							},
							{
								Key: "Expect",
								Values: []Value{
									{
										typ: stringType,
										s:   "bar",
									},
								},
							},
							{
								Key: "Hats",
								Values: []Value{
									{
										typ: numberType,
										f:   424242.42,
									},
								},
							},
						},
					},
				},
			},
			dst: &dstConf3{},
			want: &dstConf3{
				Host: []dstConf2{
					{
						KeepAlive: true,
						Expect:    []string{"foo", "bar"},
						Hats:      50, // truncated to 8bit
					},
				},
			},
		},
		{
			name:    "nil argument",
			dst:     nil,
			wantErr: true,
		},
		{
			name:    "non-pointer argument",
			dst:     int(23),
			wantErr: true,
		},
		{
			name: "block values",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("test")},
			},
			dst: &dstConf{},
			want: &dstConf{
				Args: "test",
			},
		},
		{
			name: "multiple block values",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("one"), StringValue("two")},
			},
			dst: &struct {
				Args []string
			}{},
			want: &struct {
				Args []string
			}{
				Args: []string{"one", "two"},
			},
		},
		{
			name: "block values but no Args field",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("test")},
			},
			dst:     &struct{}{},
			wantErr: true,
		},
		{
			name: "block values with type mismatch",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("not an int")},
			},
			dst: &struct {
				Args []int
			}{},
			wantErr: true,
		},
		{
			name: "multiple block values but scalar Args field",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("one"), StringValue("two")},
			},
			dst:     &dstConf{},
			wantErr: true,
		},
		{
			name: "block with children requires struct",
			src: Block{
				Key:      "Plugin",
				Children: make([]Block, 1),
			},
			dst:     stringPtr("not a struct"),
			wantErr: true,
		},
		{
			name: "error in nested block",
			src: Block{
				Key: "Plugin",
				Children: []Block{
					{
						Key:      "BlockWithErrors",
						Values:   []Value{StringValue("have string, expect int")},
						Children: make([]Block, 1),
					},
				},
			},
			dst: &struct {
				BlockWithErrors struct {
					Args int // type mismatch
				}
			}{},
			wantErr: true,
		},
		{
			name: "unexpected nested block",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("test")},
				Children: []Block{
					{
						Key:      "UnexpectedBlock",
						Children: make([]Block, 1),
					},
				},
			},
			dst: &struct {
				Args string
			}{},
			wantErr: true,
		},
		{
			name: "unmarshal list into scalar fails",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("test")},
				Children: []Block{
					{
						Key:    "ListValue",
						Values: []Value{Float64Value(23), Float64Value(64)},
					},
				},
			},
			dst: &struct {
				Args      string
				ListValue float64
			}{},
			wantErr: true,
		},
		{
			name: "unmarshal into channel fails",
			src: Block{
				Key:    "Plugin",
				Values: []Value{StringValue("test")},
				Children: []Block{
					{
						Key:    "NumberValue",
						Values: []Value{Float64Value(64)},
					},
				},
			},
			dst: &struct {
				Args        string
				NumberValue chan struct{}
			}{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.src.Unmarshal(tt.dst); (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if diff := cmp.Diff(tt.want, tt.dst); diff != "" {
				t.Errorf("%#v.Unmarshal() result differs (+got/-want):\n%s", tt.src, diff)
			}
		})
	}
}

func TestValue_Interface(t *testing.T) {
	cases := []struct {
		v    Value
		want interface{}
	}{
		{StringValue("foo"), "foo"},
		{Float64Value(42.0), 42.0},
		{BoolValue(true), true},
		{Value{}, ""}, // zero value is a string
	}

	for _, tc := range cases {
		got := tc.v.Interface()
		if !cmp.Equal(tc.want, got) {
			t.Errorf("%#v.Interface() = %v, want %v", tc.v, got, tc.want)
		}
	}
}
