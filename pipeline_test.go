package cssx

import (
	"reflect"
	"testing"
)

func TestCompileExamples(t *testing.T) {
	cases := []struct {
		input string
		want  []Op
	}{
		{
			input: ".product",
			want:  []Op{{Kind: OpSelect, Selector: ".product"}},
		},
		{
			input: ".section .item",
			want:  []Op{{Kind: OpSelect, Selector: ".section .item"}},
		},
		{
			input: ":count(.product)",
			want: []Op{
				{Kind: OpSelect, Selector: ".product"},
				{Kind: OpCall, Name: "count", Arity: 1},
			},
		},
		{
			input: ":first(section)",
			want: []Op{
				{Kind: OpSelect, Selector: "section"},
				{Kind: OpCall, Name: "first", Arity: 1},
			},
		},
		{
			input: ":text(:first(h1))",
			want: []Op{
				{Kind: OpSelect, Selector: "h1"},
				{Kind: OpCall, Name: "first", Arity: 1},
				{Kind: OpCall, Name: "text", Arity: 1},
			},
		},
		{
			input: ":attr(\"href\", :first(a.cta))",
			want: []Op{
				{Kind: OpStr, Str: "href"},
				{Kind: OpSelect, Selector: "a.cta"},
				{Kind: OpCall, Name: "first", Arity: 1},
				{Kind: OpCall, Name: "attr", Arity: 2},
			},
		},
		{
			input: ":first(a[href*=\"x,y\"])",
			want: []Op{
				{Kind: OpSelect, Selector: "a[href*=\"x,y\"]"},
				{Kind: OpCall, Name: "first", Arity: 1},
			},
		},
		{
			input: ":text(:nth(2, .section .item))",
			want: []Op{
				{Kind: OpNum, Num: 2},
				{Kind: OpSelect, Selector: ".section .item"},
				{Kind: OpCall, Name: "nth", Arity: 2},
				{Kind: OpCall, Name: "text", Arity: 1},
			},
		},
		{
			input: ".section .item >> :nth(2) >> :text()",
			want: []Op{
				{Kind: OpSelect, Selector: ".section .item"},
				{Kind: OpNum, Num: 2},
				{Kind: OpCall, Name: "nth", Arity: 1},
				{Kind: OpCall, Name: "text", Arity: 0},
			},
		},
	}

	for _, tc := range cases {
		got, err := Compile(tc.input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tc.input, err)
		}
		if !reflect.DeepEqual(got.Ops, tc.want) {
			t.Fatalf("for %q:\nwant %#v\ngot  %#v", tc.input, tc.want, got.Ops)
		}
	}
}
