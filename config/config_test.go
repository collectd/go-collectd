package config

import (
	"fmt"
	"math"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

// doubleInt implements the Unmarshaler interface to double the values on assignment.
type doubleInt int

func (di *doubleInt) UnmarshalConfig(block Block) error {
	if len(block.Values) != 1 || len(block.Children) != 0 {
		return fmt.Errorf("got %d values and %d children, want scalar value",
			len(block.Values), len(block.Children))
	}

	n, ok := block.Values[0].Number()
	if !ok {
		return fmt.Errorf("got a %T, want a number", block.Values[0].Interface())
	}

	*di = doubleInt(2.0 * n)
	return nil
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
								Key:    "KeepAlive",
								Values: Values(true),
							},
							{
								Key:    "Expect",
								Values: Values("foo"),
							},
							{
								Key:    "Expect",
								Values: Values("bar"),
							},
							{
								Key:    "Hats",
								Values: Values(424242.42),
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
								Key:    "KeepAlive",
								Values: Values(true),
							},
							{
								Key:    "Expect",
								Values: Values("foo"),
							},
							{
								Key:    "Expect",
								Values: Values("bar"),
							},
							{
								Key:    "Hats",
								Values: Values(424242.42),
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
				Values: Values("test"),
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
				Values: Values("one", "two"),
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
				Values: Values("test"),
			},
			dst:     &struct{}{},
			wantErr: true,
		},
		{
			name: "block values with type mismatch",
			src: Block{
				Key:    "Plugin",
				Values: Values("not an int"),
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
				Values: Values("one", "two"),
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
						Values:   Values("have string, expect int"),
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
				Values: Values("test"),
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
				Values: Values("test"),
				Children: []Block{
					{
						Key:    "ListValue",
						Values: Values(23, 64),
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
				Values: Values("test"),
				Children: []Block{
					{
						Key:    "NumberValue",
						Values: Values(64),
					},
				},
			},
			dst: &struct {
				Args        string
				NumberValue chan struct{}
			}{},
			wantErr: true,
		},
		{
			name: "unmarshal interface success",
			src: Block{
				Key:    "Plugin",
				Values: Values("test"),
				Children: []Block{
					{
						Key:    "Double",
						Values: Values(64),
					},
				},
			},
			dst: &struct {
				Args   string
				Double doubleInt
			}{},
			want: &struct {
				Args   string
				Double doubleInt
			}{
				Args:   "test",
				Double: doubleInt(128),
			},
		},
		{
			name: "unmarshal interface failure",
			src: Block{
				Key:    "Plugin",
				Values: Values("test"),
				Children: []Block{
					{
						Key:    "Double",
						Values: Values("not a number", 64),
					},
				},
			},
			dst: &struct {
				Args   string
				Double doubleInt
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

func TestValues(t *testing.T) {
	cases := []struct {
		in   interface{}
		want Value
	}{
		// exact matches
		{nil, Float64Value(math.NaN())},
		{"foo", StringValue("foo")},
		{[]byte("byte array"), StringValue("byte array")},
		{true, BoolValue(true)},
		{float64(42.11), Float64Value(42.11)},
		// convertible to float64
		{float32(12.25), Float64Value(12.25)},
		{int(0x1F622), Float64Value(128546)},
		{uint64(0x1F61F), Float64Value(128543)},
		// not convertiable to float64
		{complex(4, 1), StringValue("(4+1i)")},
		{struct{}{}, StringValue("{}")},
		{map[string]int{"answer": 42}, StringValue("map[answer:42]")},
		{[]int{1, 2, 3}, StringValue("[1 2 3]")},
	}

	opts := []cmp.Option{
		cmp.AllowUnexported(Value{}),
		cmpopts.EquateNaNs(),
	}
	for _, tc := range cases {
		got := Values(tc.in)
		want := []Value{tc.want}

		if !cmp.Equal(want, got, opts...) {
			t.Errorf("Values(%#v) = %v, want %v", tc.in, got, want)
		}
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

func TestBlock_Merge(t *testing.T) {
	makeBlock := func(key, value string, children []Block) Block {
		return Block{
			Key:      key,
			Values:   Values(value),
			Children: children,
		}
	}

	makeChildren := func(names ...string) []Block {
		var ret []Block
		for _, n := range names {
			ret = append(ret, makeBlock(n, "value", nil))
		}
		return ret
	}

	cases := []struct {
		name     string
		in0, in1 Block
		want     Block
		wantErr  bool
	}{
		{
			name: "success",
			in0:  makeBlock("Plugin", "test", makeChildren("foo")),
			in1:  makeBlock("Plugin", "test", makeChildren("bar")),
			want: makeBlock("Plugin", "test", makeChildren("foo", "bar")),
		},
		{
			name: "destination without children",
			in0:  makeBlock("Plugin", "test", nil),
			in1:  makeBlock("Plugin", "test", makeChildren("bar")),
			want: makeBlock("Plugin", "test", makeChildren("bar")),
		},
		{
			name: "source without children",
			in0:  makeBlock("Plugin", "test", makeChildren("foo")),
			in1:  makeBlock("Plugin", "test", nil),
			want: makeBlock("Plugin", "test", makeChildren("foo")),
		},
		{
			name: "source and destination without children",
			in0:  makeBlock("Plugin", "test", nil),
			in1:  makeBlock("Plugin", "test", nil),
			want: makeBlock("Plugin", "test", nil),
		},
		{
			name: "merge into zero value",
			in0:  Block{},
			in1:  makeBlock("Plugin", "test", makeChildren("foo")),
			want: makeBlock("Plugin", "test", makeChildren("foo")),
		},
		{
			name:    "key mismatch",
			in0:     makeBlock("Plugin", "test", nil),
			in1:     makeBlock("SomethingElse", "test", nil),
			wantErr: true,
		},
		{
			name:    "value mismatch",
			in0:     makeBlock("Plugin", "test", nil),
			in1:     makeBlock("Plugin", "prod", nil),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("block = %#v", tc.in0)
			err := tc.in0.Merge(tc.in1)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("block.Merge() = %v, want error %v", err, tc.wantErr)
			}
			if tc.wantErr {
				return
			}

			if diff := cmp.Diff(tc.want, tc.in0, cmp.AllowUnexported(Value{})); diff != "" {
				t.Errorf("other block = %#v", tc.in1)
				t.Errorf("block.Merge() differd (+got/-want)\n%s", diff)
			}
		})
	}
}
