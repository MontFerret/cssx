package cssx

// AST is the root of a parsed cssx expression.
type AST struct {
	Expr Expr
}

// Expr is any expression node.
type Expr interface {
	isExpr()
	Pos() int
}

// SelectorExpr is a raw CSS selector.
type SelectorExpr struct {
	Raw string
	pos int
}

func (*SelectorExpr) isExpr()    {}
func (s *SelectorExpr) Pos() int { return s.pos }

// CallExpr represents a pseudo call like :name(args...).
type CallExpr struct {
	Name string
	Args []Expr
	pos  int
}

func (*CallExpr) isExpr()    {}
func (c *CallExpr) Pos() int { return c.pos }

// StringLit is a quoted string literal.
type StringLit struct {
	Value string
	pos   int
}

func (*StringLit) isExpr()    {}
func (s *StringLit) Pos() int { return s.pos }

// NumberLit is a numeric literal.
type NumberLit struct {
	Value float64
	pos   int
}

func (*NumberLit) isExpr()    {}
func (n *NumberLit) Pos() int { return n.pos }

// PipelineExpr represents a base selector with a chain of calls (pipeline sugar).
type PipelineExpr struct {
	Base  *SelectorExpr
	Calls []*CallExpr
	pos   int
}

func (*PipelineExpr) isExpr()    {}
func (p *PipelineExpr) Pos() int { return p.pos }
