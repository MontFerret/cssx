package cssx

import "testing"

var benchInputs = []struct {
	name  string
	input string
}{
	{"plain_selector", ".section .item"},
	{"single_pseudo", ":count(.product)"},
	{"nested", ":text(:first(h1))"},
	{"attr_nested", ":attr(\"href\", :first(a.cta))"},
	{"comma_in_attr", ":first(a[href*=\"x,y\"])"},
	{"mixed", ":text(:nth(2, .section .item))"},
	{"pipeline_sugar", ".section .item >> :nth(2) >> :text()"},
}

func BenchmarkParse(b *testing.B) {
	for _, tc := range benchInputs {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Parse(tc.input); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkBuildPipeline(b *testing.B) {
	asts := make([]AST, len(benchInputs))
	for i, tc := range benchInputs {
		ast, err := Parse(tc.input)
		if err != nil {
			b.Fatalf("parse failed for %q: %v", tc.input, err)
		}
		asts[i] = ast
	}
	b.ResetTimer()

	for i, tc := range benchInputs {
		ast := asts[i]
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := BuildPipeline(ast); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkCompile(b *testing.B) {
	for _, tc := range benchInputs {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := Compile(tc.input); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
