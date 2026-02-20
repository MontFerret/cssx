package cssx

import "fmt"

// OpKind identifies the kind of pipeline operation.
type OpKind int

const (
	OpSelect OpKind = iota
	OpCall
	OpStr
	OpNum
)

func (k OpKind) String() string {
	switch k {
	case OpSelect:
		return "Select"
	case OpCall:
		return "Call"
	case OpStr:
		return "Str"
	case OpNum:
		return "Num"
	default:
		return fmt.Sprintf("OpKind(%d)", int(k))
	}
}

// Op is a single pipeline operation in postfix order.
type Op struct {
	Kind     OpKind
	Selector string
	Name     string
	Arity    int
	Str      string
	Num      float64
}

// Pipeline is a linear postfix pipeline representation.
type Pipeline struct {
	Ops []Op
}

// BuildPipeline compiles an AST into a postfix pipeline IR.
func BuildPipeline(ast AST) (Pipeline, error) {
	if ast.Expr == nil {
		return Pipeline{}, &ParseError{Msg: "empty AST", Pos: 0}
	}
	var p Pipeline
	if err := buildExpr(ast.Expr, &p); err != nil {
		return Pipeline{}, err
	}
	return p, nil
}

func buildExpr(expr Expr, p *Pipeline) error {
	switch e := expr.(type) {
	case *SelectorExpr:
		p.Ops = append(p.Ops, Op{Kind: OpSelect, Selector: e.Raw})
	case *StringLit:
		p.Ops = append(p.Ops, Op{Kind: OpStr, Str: e.Value})
	case *NumberLit:
		p.Ops = append(p.Ops, Op{Kind: OpNum, Num: e.Value})
	case *CallExpr:
		for _, arg := range e.Args {
			if err := buildExpr(arg, p); err != nil {
				return err
			}
		}
		p.Ops = append(p.Ops, Op{Kind: OpCall, Name: e.Name, Arity: len(e.Args)})
	case *PipelineExpr:
		if e.Base == nil {
			return &ParseError{Msg: "pipeline missing base selector", Pos: e.Pos()}
		}
		if err := buildExpr(e.Base, p); err != nil {
			return err
		}
		for _, call := range e.Calls {
			if err := buildExpr(call, p); err != nil {
				return err
			}
		}
	default:
		return &ParseError{Msg: "unknown expression type", Pos: 0}
	}
	return nil
}

// Compile parses input and builds a pipeline.
func Compile(input string) (Pipeline, error) {
	ast, err := Parse(input)
	if err != nil {
		return Pipeline{}, err
	}
	return BuildPipeline(ast)
}
