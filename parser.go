package cssx

import (
	"strconv"
	"strings"
)

// Parse parses input into an AST.
func Parse(input string) (AST, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return AST{}, &ParseError{Msg: "empty input", Pos: 0}
	}

	// Pipeline sugar: base >> :call() >> :call()
	if segments, ok := splitPipelineSegments(input); ok {
		ast, err := parsePipelineSegments(segments)
		if err != nil {
			return AST{}, err
		}
		return ast, nil
	}

	// Expression mode or plain selector mode.
	p := newParser(input, 0)
	p.skipWS()
	if p.eof() {
		return AST{}, &ParseError{Msg: "empty input", Pos: 0}
	}
	if p.peek() == ':' {
		expr, err := p.parseExpr()
		if err != nil {
			return AST{}, err
		}
		p.skipWS()
		if !p.eof() {
			return AST{}, p.errf("unexpected trailing input", p.pos)
		}
		return AST{Expr: expr}, nil
	}

	// Plain selector mode.
	return AST{Expr: &SelectorExpr{Raw: strings.TrimSpace(input), pos: 0}}, nil
}

// ParseToAST is an alias for Parse.
func ParseToAST(input string) (AST, error) {
	return Parse(input)
}

type parser struct {
	input string
	pos   int
	base  int
}

func newParser(input string, base int) *parser {
	return &parser{input: input, pos: 0, base: base}
}

func (p *parser) eof() bool {
	return p.pos >= len(p.input)
}

func (p *parser) peek() byte {
	if p.eof() {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) next() byte {
	if p.eof() {
		return 0
	}
	b := p.input[p.pos]
	p.pos++
	return b
}

func (p *parser) skipWS() {
	for !p.eof() && isWS(p.peek()) {
		p.pos++
	}
}

func (p *parser) errf(msg string, pos int) error {
	return &ParseError{Msg: msg, Pos: p.base + pos}
}

func (p *parser) parseExpr() (Expr, error) {
	p.skipWS()
	if p.eof() {
		return nil, p.errf("unexpected end of input", p.pos)
	}
	ch := p.peek()
	if ch == ',' || ch == ')' {
		return nil, p.errf("expected expression", p.pos)
	}
	if ch == ':' {
		return p.parseCall()
	}
	if ch == '"' || ch == '\'' {
		return p.parseString()
	}
	if isDigit(ch) || (ch == '-' && p.hasNumberStart()) {
		return p.parseNumber()
	}
	return p.parseSelector()
}

func (p *parser) parseCall() (*CallExpr, error) {
	start := p.pos
	if p.peek() != ':' {
		return nil, p.errf("expected ':'", p.pos)
	}
	p.pos++
	p.skipWS()
	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	p.skipWS()
	if p.eof() || p.peek() != '(' {
		return nil, p.errf("expected '('", p.pos)
	}
	p.pos++
	p.skipWS()
	args := []Expr{}
	if !p.eof() && p.peek() == ')' {
		p.pos++
		return &CallExpr{Name: name, Args: args, pos: p.base + start}, nil
	}
	for {
		arg, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		p.skipWS()
		if p.eof() {
			return nil, p.errf("expected ')'", p.pos)
		}
		switch p.peek() {
		case ',':
			p.pos++
			p.skipWS()
			if !p.eof() && p.peek() == ')' {
				return nil, p.errf("expected expression", p.pos)
			}
			continue
		case ')':
			p.pos++
			return &CallExpr{Name: name, Args: args, pos: p.base + start}, nil
		default:
			return nil, p.errf("expected ',' or ')'", p.pos)
		}
	}
}

func (p *parser) parseIdent() (string, error) {
	p.skipWS()
	if p.eof() {
		return "", p.errf("expected identifier", p.pos)
	}
	start := p.pos
	ch := p.peek()
	if !isIdentStart(ch) {
		return "", p.errf("expected identifier", p.pos)
	}
	p.pos++
	for !p.eof() && isIdentPart(p.peek()) {
		p.pos++
	}
	return p.input[start:p.pos], nil
}

func (p *parser) parseString() (*StringLit, error) {
	start := p.pos
	quote := p.next()
	var sb strings.Builder
	for !p.eof() {
		ch := p.next()
		switch ch {
		case '\\':
			if p.eof() {
				return nil, p.errf("unterminated string", p.pos)
			}
			sb.WriteByte(p.next())
		case quote:
			return &StringLit{Value: sb.String(), pos: p.base + start}, nil
		default:
			sb.WriteByte(ch)
		}
	}
	return nil, p.errf("unterminated string", p.pos)
}

func (p *parser) parseNumber() (*NumberLit, error) {
	start := p.pos
	if p.peek() == '-' {
		p.pos++
	}
	digitsBefore := p.consumeDigits()
	digitsAfter := 0
	if !p.eof() && p.peek() == '.' {
		p.pos++
		digitsAfter = p.consumeDigits()
	}
	if digitsBefore == 0 && digitsAfter == 0 {
		return nil, p.errf("invalid number", start)
	}
	txt := p.input[start:p.pos]
	val, err := strconv.ParseFloat(txt, 64)
	if err != nil {
		return nil, p.errf("invalid number", start)
	}
	return &NumberLit{Value: val, pos: p.base + start}, nil
}

func (p *parser) parseSelector() (*SelectorExpr, error) {
	start := p.pos
	inString := byte(0)
	escape := false
	bracketDepth := 0
	parenDepth := 0
	for !p.eof() {
		ch := p.peek()
		if inString != 0 {
			p.pos++
			if escape {
				escape = false
				continue
			}
			if ch == '\\' {
				escape = true
				continue
			}
			if ch == inString {
				inString = 0
			}
			continue
		}
		switch ch {
		case '\'', '"':
			inString = ch
			p.pos++
		case '[':
			bracketDepth++
			p.pos++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			p.pos++
		case '(':
			parenDepth++
			p.pos++
		case ')':
			if bracketDepth == 0 && parenDepth == 0 {
				goto done
			}
			if parenDepth > 0 {
				parenDepth--
			}
			p.pos++
		case ',':
			if bracketDepth == 0 && parenDepth == 0 {
				goto done
			}
			p.pos++
		default:
			p.pos++
		}
	}

done:
	raw := strings.TrimSpace(p.input[start:p.pos])
	if raw == "" {
		return nil, p.errf("expected selector", start)
	}
	return &SelectorExpr{Raw: raw, pos: p.base + start}, nil
}

func (p *parser) consumeDigits() int {
	count := 0
	for !p.eof() && isDigit(p.peek()) {
		p.pos++
		count++
	}
	return count
}

func (p *parser) hasNumberStart() bool {
	if p.peek() != '-' {
		return false
	}
	if p.pos+1 >= len(p.input) {
		return false
	}
	n := p.input[p.pos+1]
	return isDigit(n) || n == '.'
}

func isWS(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r', '\f':
		return true
	default:
		return false
	}
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || isDigit(ch) || ch == '-'
}

type segment struct {
	text  string
	start int
}

func splitPipelineSegments(input string) ([]segment, bool) {
	inString := byte(0)
	escape := false
	bracketDepth := 0
	parenDepth := 0
	start := 0
	found := false
	segments := []segment{}

	for i := 0; i < len(input); {
		ch := input[i]
		if inString != 0 {
			if escape {
				escape = false
				i++
				continue
			}
			if ch == '\\' {
				escape = true
				i++
				continue
			}
			if ch == inString {
				inString = 0
			}
			i++
			continue
		}

		switch ch {
		case '\'', '"':
			inString = ch
			i++
		case '[':
			bracketDepth++
			i++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			i++
		case '(':
			parenDepth++
			i++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			i++
		case '>':
			if bracketDepth == 0 && parenDepth == 0 && i+1 < len(input) && input[i+1] == '>' {
				found = true
				segments = append(segments, segment{text: input[start:i], start: start})
				i += 2
				start = i
				continue
			}
			i++
		default:
			i++
		}
	}

	if found {
		segments = append(segments, segment{text: input[start:], start: start})
	}
	return segments, found
}

func parsePipelineSegments(segments []segment) (AST, error) {
	if len(segments) == 0 {
		return AST{}, &ParseError{Msg: "empty pipeline", Pos: 0}
	}

	baseSeg, ok := trimSegment(segments[0])
	if !ok {
		return AST{}, &ParseError{Msg: "pipeline base selector is empty", Pos: segments[0].start}
	}
	if strings.HasPrefix(baseSeg.text, ":") {
		return AST{}, &ParseError{Msg: "pipeline base must be a selector", Pos: baseSeg.start}
	}
	base := &SelectorExpr{Raw: baseSeg.text, pos: baseSeg.start}

	calls := []*CallExpr{}
	for _, seg := range segments[1:] {
		trimmed, ok := trimSegment(seg)
		if !ok {
			return AST{}, &ParseError{Msg: "pipeline stage is empty", Pos: seg.start}
		}
		p := newParser(trimmed.text, trimmed.start)
		p.skipWS()
		if p.eof() || p.peek() != ':' {
			return AST{}, &ParseError{Msg: "pipeline stage must be a call", Pos: trimmed.start}
		}
		call, err := p.parseCall()
		if err != nil {
			return AST{}, err
		}
		p.skipWS()
		if !p.eof() {
			return AST{}, p.errf("unexpected trailing input", p.pos)
		}
		calls = append(calls, call)
	}

	return AST{Expr: &PipelineExpr{Base: base, Calls: calls, pos: baseSeg.start}}, nil
}

func trimSegment(seg segment) (segment, bool) {
	left := 0
	right := len(seg.text)
	for left < right && isWS(seg.text[left]) {
		left++
	}
	for right > left && isWS(seg.text[right-1]) {
		right--
	}
	trimmed := seg.text[left:right]
	if trimmed == "" {
		return segment{}, false
	}
	return segment{text: trimmed, start: seg.start + left}, true
}
