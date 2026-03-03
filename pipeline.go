package cssx

import "fmt"

// OpKind identifies the kind of pipeline operation.
type OpKind int

const (
	OpSelect OpKind = iota
	OpCall
)

func (k OpKind) String() string {
	switch k {
	case OpSelect:
		return "Select"
	case OpCall:
		return "Call"
	default:
		return fmt.Sprintf("OpKind(%d)", int(k))
	}
}

// CallArgKind identifies the kind of a call literal argument.
type CallArgKind int

const (
	CallArgString CallArgKind = iota
	CallArgNumber
)

// CallArg is an embedded literal argument for a call operation.
type CallArg struct {
	Kind CallArgKind
	Str  string
	Num  float64
}

// Op is a single pipeline operation in postfix order.
type Op struct {
	Kind     OpKind
	Selector string
	Name     string
	Arity    int
	Args     []CallArg
}

// Pipeline is a linear postfix pipeline representation.
type Pipeline struct {
	Ops []Op
}

// BuildPipeline compiles an AST into a postfix pipeline IR.
func BuildPipeline(ast AST) (Pipeline, error) {
	if ast.Expr == nil {
		return Pipeline{}, &ParseError{Message: "empty AST", Pos: 0}
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
		return &ParseError{Message: "literal values are only allowed as call arguments", Pos: e.Pos()}
	case *NumberLit:
		return &ParseError{Message: "literal values are only allowed as call arguments", Pos: e.Pos()}
	case *CallExpr:
		return buildCall(e, p)
	case *PipelineExpr:
		if e.Base == nil {
			return &ParseError{Message: "pipeline missing base selector", Pos: e.Pos()}
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
		return &ParseError{Message: "unknown expression type", Pos: 0}
	}
	return nil
}

func buildCall(call *CallExpr, p *Pipeline) error {
	op := Op{Kind: OpCall, Name: call.Name}
	seenExpr := false
	for _, arg := range call.Args {
		switch a := arg.(type) {
		case *StringLit:
			if seenExpr {
				return &ParseError{Message: "literal args must precede expression args", Pos: a.Pos()}
			}
			op.Args = append(op.Args, CallArg{Kind: CallArgString, Str: a.Value})
		case *NumberLit:
			if seenExpr {
				return &ParseError{Message: "literal args must precede expression args", Pos: a.Pos()}
			}
			op.Args = append(op.Args, CallArg{Kind: CallArgNumber, Num: a.Value})
		default:
			seenExpr = true
			if err := buildExpr(arg, p); err != nil {
				return err
			}
			op.Arity++
		}
	}
	p.Ops = append(p.Ops, op)
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
