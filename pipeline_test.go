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
				{Kind: OpSelect, Selector: "a.cta"},
				{Kind: OpCall, Name: "first", Arity: 1},
				{Kind: OpCall, Name: "attr", Arity: 1, Args: []CallArg{{Kind: CallArgString, Str: "href"}}},
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
				{Kind: OpSelect, Selector: ".section .item"},
				{Kind: OpCall, Name: "nth", Arity: 1, Args: []CallArg{{Kind: CallArgNumber, Num: 2}}},
				{Kind: OpCall, Name: "text", Arity: 1},
			},
		},
		{
			input: ".section .item >> :nth(2) >> :text()",
			want: []Op{
				{Kind: OpSelect, Selector: ".section .item"},
				{Kind: OpCall, Name: "nth", Arity: 0, Args: []CallArg{{Kind: CallArgNumber, Num: 2}}},
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

func TestCompileLiteralOrderValidation(t *testing.T) {
	cases := []string{
		":foo(:bar(a), \"x\")",
		":foo(1, :bar(a), 2)",
	}

	for _, input := range cases {
		_, err := Compile(input)
		if err == nil {
			t.Fatalf("expected compile error for %q", input)
		}
		parseErr, ok := err.(*ParseError)
		if !ok {
			t.Fatalf("expected ParseError for %q, got %T", input, err)
		}
		if parseErr.Message != "literal args must precede expression args" {
			t.Fatalf("unexpected message for %q: %q", input, parseErr.Message)
		}
	}
}

func TestCompileLiteralBeforeExpressionSuccess(t *testing.T) {
	got, err := Compile(":foo(\"x\", 2, :bar(a))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []Op{
		{Kind: OpSelect, Selector: "a"},
		{Kind: OpCall, Name: "bar", Arity: 1},
		{Kind: OpCall, Name: "foo", Arity: 1, Args: []CallArg{{Kind: CallArgString, Str: "x"}, {Kind: CallArgNumber, Num: 2}}},
	}

	if !reflect.DeepEqual(got.Ops, want) {
		t.Fatalf("want %#v\ngot  %#v", want, got.Ops)
	}
}
