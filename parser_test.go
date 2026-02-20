package cssx

import "testing"

func TestParsePlainSelector(t *testing.T) {
	ast, err := Parse(".product")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	sel, ok := ast.Expr.(*SelectorExpr)
	if !ok {
		t.Fatalf("expected SelectorExpr, got %T", ast.Expr)
	}
	if sel.Raw != ".product" {
		t.Fatalf("expected selector '.product', got %q", sel.Raw)
	}
}

func TestParseExpression(t *testing.T) {
	ast, err := Parse(":text(:first(h1))")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	call, ok := ast.Expr.(*CallExpr)
	if !ok {
		t.Fatalf("expected CallExpr, got %T", ast.Expr)
	}
	if call.Name != "text" || len(call.Args) != 1 {
		t.Fatalf("unexpected call: %#v", call)
	}
}

func TestParsePipelineSugar(t *testing.T) {
	ast, err := Parse(".section .item >> :nth(2) >> :text()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pipe, ok := ast.Expr.(*PipelineExpr)
	if !ok {
		t.Fatalf("expected PipelineExpr, got %T", ast.Expr)
	}
	if pipe.Base == nil || pipe.Base.Raw != ".section .item" {
		t.Fatalf("unexpected base selector: %#v", pipe.Base)
	}
	if len(pipe.Calls) != 2 {
		t.Fatalf("expected 2 calls, got %d", len(pipe.Calls))
	}
}

func TestParseErrors(t *testing.T) {
	cases := []string{
		":",
		":text(",
		":text(:first(h1)",
		":attr(\"href\" :first(a))",
		"  ",
		".section >> text()",
	}
	for _, input := range cases {
		if _, err := Parse(input); err == nil {
			t.Fatalf("expected error for %q", input)
		}
	}
}
